import os
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

from database import init_database, close_database
from routes import analytics, health

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize OpenTelemetry
def init_tracing():
    base_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://opentelemetry-collector.openchoreo-observability-plane.svc.cluster.local:4318")
    logger.info(f"Setting up OpenTelemetry tracing to: {base_endpoint}")

    # Ensure endpoint has the correct path for OTLP HTTP
    endpoint = base_endpoint
    if not endpoint.endswith("/v1/traces"):
        endpoint = f"{endpoint}/v1/traces"

    resource = Resource(attributes={
        ResourceAttributes.SERVICE_NAME: "analytics-service"
    })

    provider = TracerProvider(resource=resource)
    processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint))
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)

    logger.info(f"OpenTelemetry tracing successfully initialized (service=analytics-service, endpoint={endpoint})")

# Initialize tracing
logger.info("Initializing OpenTelemetry tracing...")
init_tracing()


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    try:
        init_database()
    except Exception as e:
        logger.error(f"Failed to initialize postgres: {e}")
        logger.warning("Service starting without database connection. Database operations will fail until connection is established.")

    yield

    # Shutdown
    close_database()


app = FastAPI(title="Analytics Service", lifespan=lifespan)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# OpenTelemetry auto-instrumentation for FastAPI
FastAPIInstrumentor.instrument_app(app)

# Include routers
app.include_router(health.router)
app.include_router(analytics.router)


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "7544"))
    uvicorn.run(app, host="0.0.0.0", port=port)
