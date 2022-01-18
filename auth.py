import base64
import json
from typing import Optional

from fastapi import FastAPI, Header, Response, status

app = FastAPI()

USERS_REPO = {
    "user1": {
        "username": "foo",
        "role": "admin",
        "permissions": [
            "private/view",
            "private/edit",
            "private/list",
        ],
    },
    "user2": {
        "username": "bar",
        "role": "user",
        "permissions": [
            "private/view",
            "private/list",
        ],
    },
}


@app.get("/auth")
async def root(
    response: Response,
    authorization: Optional[str] = Header(None),
):
    # check authorization
    user_data = USERS_REPO.get(authorization, None)
    if user_data is None:
        response.status_code = status.HTTP_401_UNAUTHORIZED
        return

    json_user_data = json.dumps(user_data)

    encoded_user_data = base64.b64encode(json_user_data.encode())
    response.headers["X-User-Data"] = encoded_user_data.decode()
    return
