from fastapi import APIRouter, Request

router = APIRouter()


@router.get("/healthz", tags=["system"])
async def healthz(request: Request) -> dict[str, str]:
    settings = request.app.state.container.settings
    return {
        "status": "ok",
        "service": settings.app_name,
        "environment": settings.app_env,
    }
