from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.api.routes import router
from app.core.container import AppContainer
from app.core.config import get_settings
from app.core.logging import setup_logging


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = get_settings()
    container = AppContainer(settings)
    app.state.container = container
    await container.startup()
    try:
        yield
    finally:
        await container.shutdown()


def create_app() -> FastAPI:
    settings = get_settings()
    setup_logging()
    app = FastAPI(title=settings.app_name, lifespan=lifespan)
    app.include_router(router)
    return app


app = create_app()
