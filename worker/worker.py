import redis
import json
import time
import os
import pickle
from minio import Minio, S3Error
from io import BytesIO
import pandas as pd
import re
from data_drift import check_data_drift
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score
import shutil
import traceback
import joblib
import numpy as np

RESULTS_QUEUE = "results_queue"
QUEUE_NAME = "drift_check_queue"


def connect_redis():
    while True:
        try:
            r = redis.Redis(host="redis", port=6379, db=0)
            r.ping()
            print("connected to Redis")
            return r
        except:
            print("waiting for Redis...")
            time.sleep(2)


def get_minio_client():
    return Minio(
        os.getenv("MINIO_ENDPOINT"),
        access_key=os.getenv("MINIO_ACCESS_KEY"),
        secret_key=os.getenv("MINIO_SECRET_KEY"),
        secure=False
    )


def download_data_from_bucket(storage_path,type):
    client = get_minio_client()
    bucket = type
    os.makedirs(f"/tmp/{bucket}", exist_ok=True)

    local_path = f"/tmp/{bucket}/{storage_path.split('/')[-1]}"

    try:
        client.fget_object(bucket, storage_path, local_path)
        print(f"MINIO - downloaded {storage_path}")
        return local_path

    except S3Error as e:
        if e.code == "NoSuchKey":
            print(f"[WARN] File not found in MinIO: {storage_path}")
            return None
        else:
            print(f"[ERROR] MinIO error: {e}")
            return None

    except Exception as e:
        print(f"[ERROR] Unexpected error downloading {storage_path}: {e}")
        return None


def upload_bucket(model_path, model_obj, bucket):

    client = get_minio_client()


    match = re.search(
        r"v(\d+)\.(pkl|csv)$",
        model_path
    )

    if not match:
        raise ValueError(
            "invalid file path format"
        )

    current_version = int(match.group(1))
    new_version = current_version + 1

    extension = (
        "pkl"
        if bucket == "models"
        else "csv"
    )

    new_model_path = re.sub(
        r"v\d+\.(pkl|csv)$",
        f"v{new_version}.{extension}",
        model_path,
    )

    buffer = BytesIO()

    if bucket == "models":

        pickle.dump(model_obj, buffer)

        content_type = (
            "application/octet-stream"
        )

    else:
        if not isinstance(
            model_obj,
            pd.DataFrame
        ):
            raise ValueError(
                "For non-model buckets, "
                "model_obj must be a pandas DataFrame"
            )

        csv_bytes = model_obj.to_csv(
            index=False
        ).encode("utf-8")

        buffer.write(csv_bytes)

        content_type = "text/csv"

    buffer.seek(0)

    client.put_object(
        bucket_name=bucket,
        object_name=new_model_path,
        data=buffer,
        length=buffer.getbuffer().nbytes,
        content_type=content_type,
    )

    print(
        f"MINIO - uploaded new version: "
        f"{new_model_path}"
    )

    return new_model_path

def load_model(path):
    try:
        return joblib.load(path)
    except Exception as e:
        raise ValueError(f"Failed to load model with joblib: {e}")

def download_from_github(datasetlink):
    df = pd.read_csv(datasetlink)
    return df

def process_task(task):
    storage_path = task["storage_path"]
    datasetlink = task["dataset_link"]
    datasetpath = task["dataset_path"]
    projectid = task["project_id"]


    results = {
        "status": "success",
        "data_drift_detected": False,
        "error_message": None,
        "traceback": None
    }
    try:
        model_path = download_data_from_bucket(storage_path,"models")
        model = load_model(model_path)

        old_data = pd.read_csv(
        download_data_from_bucket(datasetpath, "datasets")
        )

        current_data = download_from_github(datasetlink)

        results = check_data_drift(
        old_data,
        current_data
        )

        try:
            X_current = current_data.drop("target", axis=1)
            y_current = current_data["target"]
            current_predictions = model.predict(X_current)
            results["accuracy"] = accuracy_score(y_current, current_predictions)
        except Exception as eval_err:
            print(f"[WARN] Failed to calculate current accuracy: {eval_err}")
            results["accuracy"] = None

        if  results["data_drift_detected"]:
            print("Drift detected! Retraining model...")
            X = current_data.drop("target", axis=1)
            y = current_data["target"]
            X_train, X_test, y_train, y_test = train_test_split(
            X,
            y,
            test_size=0.2,
            random_state=42
            )
            model.fit(X_train, y_train)
            predictions = model.predict(X_test)

            accuracy = accuracy_score(y_test, predictions)
            results["accuracy"] = accuracy
            results["datasetpath"] = upload_bucket(datasetpath,current_data,"datasets")
            results["storage_path"] = upload_bucket(storage_path,model,"models")
            print("finisehd training model")

    except Exception as e:
        error_stack = traceback.format_exc()

        results["status"] = "failed"
        results["error_message"] = str(e)
        results["traceback"] = error_stack

        print(f"\n\033[91m[TASK FAILED] {e}")
        print(f"--- TRACEBACK ---\n{error_stack}\033[0m\n")
    finally:
        print("Cleaning up temporary files...")
        shutil.rmtree("/tmp/models", ignore_errors=True)
        shutil.rmtree("/tmp/datasets", ignore_errors=True)

    results["project_id"] = str(projectid)
    return results

def json_converter(obj):
    if isinstance(obj, np.bool_):
        return bool(obj)

    if isinstance(obj, np.integer):
        return int(obj)

    if isinstance(obj, np.floating):
        return float(obj)

    raise TypeError(f"Object of type {type(obj)} is not JSON serializable")
def main():
    r = connect_redis()

    print("worker ready")

    while True:
        task_data = r.brpop(QUEUE_NAME, timeout=5)

        if not task_data:
            continue

        _, task_json = task_data

        try:
            task = json.loads(task_json.decode("utf-8"))
            results = process_task(task)

            r.lpush(
                RESULTS_QUEUE,
                json.dumps(results, default=json_converter)
            )

            print(
                f"results pushed to {RESULTS_QUEUE}")
        except Exception as e:
            print("[ERROR]", e)


if __name__ == "__main__":
    main()