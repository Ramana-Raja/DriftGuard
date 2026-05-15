import redis
import json
import time
import os
import pickle
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
from azure.storage.blob import BlobServiceClient


RESULTS_QUEUE = "results_queue"
QUEUE_NAME = "drift_check_queue"


def connect_redis():
    while True:
        try:
            r = redis.Redis(
                host=os.getenv("REDIS_HOST", "localhost"),
                port=6379,
                db=0
            )
            r.ping()
            print("connected to Redis")
            return r
        except:
            print("waiting for Redis...")
            time.sleep(2)


def get_blob_service_client():
    account = os.getenv("AZURE_STORAGE_ACCOUNT")
    key = os.getenv("AZURE_STORAGE_KEY")

    connection_string = (
        f"DefaultEndpointsProtocol=https;"
        f"AccountName={account};"
        f"AccountKey={key};"
        f"EndpointSuffix=core.windows.net"
    )

    return BlobServiceClient.from_connection_string(
        connection_string
    )

def download_data_from_bucket(storage_path, container_name):
    client = get_blob_service_client()

    if "blob.core.windows.net" in storage_path:
        search_str = f"/{container_name}/"
        if search_str in storage_path:
            storage_path = storage_path.split(search_str)[-1]

    print(f"DEBUG: Cleaned blob path for Azure request: {storage_path}")

    blob_client = client.get_blob_client(
        container=container_name,
        blob=storage_path
    )

    os.makedirs(f"/tmp/{container_name}", exist_ok=True)
    local_path = f"/tmp/{container_name}/{storage_path.split('/')[-1]}"

    try:
        with open(local_path, "wb") as file:
            download_stream = blob_client.download_blob()
            file.write(download_stream.readall())

        print(f"AZURE BLOB - downloaded {storage_path}")
        return local_path

    except Exception as e:
        print(f"[ERROR] Blob download failed for path '{storage_path}': {e}")
        raise e

def upload_bucket(model_path, model_obj, container_name):
    client = get_blob_service_client()

    if "blob.core.windows.net" in model_path:
        search_str = f"/{container_name}/"
        if search_str in model_path:
            model_path = model_path.split(search_str)[-1]

    print(f"DEBUG: Cleaned base path for versioning: {model_path}")

    match = re.search(
        r"v(\d+)\.(pkl|csv)$",
        model_path
    )

    if not match:
        raise ValueError(f"Invalid file path format for versioning: {model_path}")

    current_version = int(match.group(1))
    new_version = current_version + 1

    extension = (
        "pkl"
        if container_name == "models"
        else "csv"
    )

    new_model_path = re.sub(
        r"v\d+\.(pkl|csv)$",
        f"v{new_version}.{extension}",
        model_path,
    )

    buffer = BytesIO()

    if container_name == "models":
        pickle.dump(model_obj, buffer)
    else:
        csv_bytes = model_obj.to_csv(index=False).encode("utf-8")
        buffer.write(csv_bytes)

    buffer.seek(0)

    blob_client = client.get_blob_client(
        container=container_name,
        blob=new_model_path
    )

    blob_client.upload_blob(
        buffer,
        overwrite=True
    )

    print(f"AZURE BLOB - successfully uploaded version update: {new_model_path}")

    account = os.getenv("AZURE_STORAGE_ACCOUNT")
    return f"https://{account}.blob.core.windows.net/{container_name}/{new_model_path}"

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