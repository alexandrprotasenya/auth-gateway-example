import base64
import json

from fastapi import Depends, FastAPI, Header, HTTPException, status

app = FastAPI()


async def decode_user(x_user_data: str = Header(...)):
    user_data = json.loads(base64.b64decode(x_user_data.encode()).decode())
    return user_data


def check_permission(permission):
    def allow(user_data: dict = Depends(decode_user)):
        user_permissions = user_data.get("permissions", [])
        if permission not in user_permissions:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN, detail="Forbiden"
            )
        return user_data

    return allow


@app.get("/view", dependencies=[Depends(check_permission("private/view"))])
async def view_resource(user_data: dict = Depends(decode_user)):
    return {"message": "view allowed", "user": user_data}


@app.get("/edit", dependencies=[Depends(check_permission("private/edit"))])
async def edit_resource():
    return {"message": "edit allowed"}


@app.get("/list", dependencies=[Depends(check_permission("private/list"))])
async def list_resources():
    return {"message": "list allowed"}
