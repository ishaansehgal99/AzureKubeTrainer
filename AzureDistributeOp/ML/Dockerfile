FROM nvcr.io/nvidia/pytorch:23.06-py3
WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY training.py .
