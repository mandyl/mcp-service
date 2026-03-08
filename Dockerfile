FROM python:3.11-slim

WORKDIR /app

# 使用国内 pip 镜像加速依赖安装
COPY requirements.txt .
RUN pip install --no-cache-dir -i https://pypi.tuna.tsinghua.edu.cn/simple -r requirements.txt

COPY main.py .

EXPOSE 8080

# Bug fix 5: 使用 gunicorn 替代 Flask dev server
CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "2", "--timeout", "60", "main:app"]
