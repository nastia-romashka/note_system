from pathlib import Path

ANY_ORIGIN = {"origins": "*"}

X_TRACE_ID = "X-B3-TraceId"
LOCATION = "Location"

EXPOSE_HEADERS = [LOCATION, X_TRACE_ID]

SERVICE_ROOT = Path(__file__).resolve().parents[1]

LOG_DIR = str(SERVICE_ROOT / "logs")

CONFIG_FILE_PATH = str(SERVICE_ROOT / "config.yaml")
