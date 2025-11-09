from fastapi import FastAPI
from time import sleep
from fastapi.middleware.cors import CORSMiddleware

app: FastAPI = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


busy_duration: int = 5  
busy: bool = False

@app.get("/")
async def read_root():
    return {"Hello": "World"}

@app.get("/health")
def health_check():
    return "ok" 

@app.get("/busy")
def isbusy():
    global busy
    return "yes" if busy else "no"

@app.post("/task")
async def process_task(task: dict = {}):
    # this function simulates long computation on the server
    sleep(busy_duration) 
    return {"status": "processed", "task": task}

