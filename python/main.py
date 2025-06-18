# validate_server.py
import bili_ticket_gt_python
from fastapi import FastAPI, HTTPException
from loguru import logger
from pydantic import BaseModel, Field
from typing import Literal

from pool_manager import GeetestPool

pool = GeetestPool()
app = FastAPI()


def do_validate(gt: str, challenge: str) -> str:
    click = bili_ticket_gt_python.ClickPy()
    return click.simple_match_retry(gt, challenge)


class GeetestRequest(BaseModel):
    type: Literal["geetest"]
    gt: str
    challenge: str


class GeetestResponse(BaseModel):
    validate_: str = Field(alias="validate")
    seccode: str


@app.post("/validate/geetest", response_model=GeetestResponse)
def validate_geetest(req: GeetestRequest):
    try:
        task_id = pool.submit(req.gt, req.challenge)
        validate, seccode, error = pool.get_result(task_id)
        if error:
            raise HTTPException(status_code=500, detail=error)
        return {"validate": validate, "seccode": seccode}
    except Exception as e:
        logger.exception(e)
        raise HTTPException(status_code=500, detail=str(e))
