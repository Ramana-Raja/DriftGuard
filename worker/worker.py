import redis
import json
import time
import os
import pickle
from minio import Minio
from io import BytesIO

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


def download_model(storage_path):
    client = get_minio_client()
    bucket = "models"

    local_path = f"/tmp/{storage_path.split('/')[-1]}"

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

def upload_model(original_path, model_obj):
    client = get_minio_client()
    bucket = "models"

    buffer = BytesIO()
    pickle.dump(model_obj, buffer)
    buffer.seek(0)

    client.put_object(
        bucket,
        original_path,
        buffer,
        length=buffer.getbuffer().nbytes,
        content_type="application/octet-stream"
    )

    print(f"MINIO - uploaded new version: {new_object_name}")

    return new_object_name
def load_model(path):
    with open(path, "rb") as f:
        return pickle.load(f)


def process_task(task):
    storage_path = task["storage_path"]

    local_file = download_model(storage_path)

    if not local_file:
        print("Skipping task - file not available")
        return

    try:
        model = load_model(local_file)
        print("model loaded successfully")
    except Exception as e:
        print(f"[ERROR] Failed to load model: {e}")

    upload_model(storage_path,model)

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
            process_task(task)
        except Exception as e:
            print("[ERROR]", e)


if __name__ == "__main__":
    main()