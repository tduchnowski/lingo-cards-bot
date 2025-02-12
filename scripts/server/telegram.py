import random
import time
from fastapi import FastAPI
from fastapi.responses import ORJSONResponse


app = FastAPI()

@app.get('/bot{token}/getUpdates')
def get_updates(token:str, response_class=ORJSONResponse):
    response = {
        "ok":True,
        "result":[
            {
                "update_id":651277849,
                "message":{
                    "message_id":22,
                    "from":{
                        "id":2144122443,
                        "is_bot":False,
                        "first_name":"Tomasz",
                        "username":"CosmicBuddy",
                        "language_code":"en"
                    },
                    "chat":{
                        "id":2144122443,
                        "first_name":"Tomasz",
                        "username":"CosmicBuddy",
                        "type":"private"
                    },
                    "date":1738845518,
                    "text":"/start",
                    "entities":[
                        {
                            "offset":0,
                            "length":6,
                            "type":"bot_command"
                        }
                    ]
                }
            },
            {
                "update_id":651277850,
                "message":{
                    "message_id":23,
                    "from":{
                        "id":2144122443,
                        "is_bot":False,
                        "first_name":"Tomasz",
                        "username":"CosmicBuddy",
                        "language_code":"en"
                    },
                    "chat":{
                        "id":2144122443,
                        "first_name":"Tomasz",
                        "username":"CosmicBuddy",
                        "type":"private"
                    },
                    "date":1738845519,"text":"/start","entities":[{"offset":0,"length":6,"type":"bot_command"}]}},
            {"update_id":651277851,
"message":{"message_id":24,"from":{"id":2144122443,"is_bot":False,"first_name":"Tomasz","username":"CosmicBuddy","language_code":"en"},"chat":{"id":2144122443,"first_name":"Tomasz","username":"CosmicBuddy","type":"private"},"date":1738845520,"text":"/start","entities":[{"offset":0,"length":6,"type":"bot_command"}]}},{"update_id":651277852,
"message":{"message_id":25,"from":{"id":2144122443,"is_bot":False,"first_name":"Tomasz","username":"CosmicBuddy","language_code":"en"},"chat":{"id":2144122443,"first_name":"Tomasz","username":"CosmicBuddy","type":"private"},"date":1738845521,"text":"/start","entities":[{"offset":0,"length":6,"type":"bot_command"}]}},{"update_id":651277853,
"message":{"message_id":26,"from":{"id":2144122443,"is_bot":False,"first_name":"Tomasz","username":"CosmicBuddy","language_code":"en"},"chat":{"id":2144122443,"first_name":"Tomasz","username":"CosmicBuddy","type":"private"},"date":1738845523,"text":"/start","entities":[{"offset":0,"length":6,"type":"bot_command"}]}}]}
    return ORJSONResponse(response)

@app.get('/bot{token}/getMe')
def get_me(token:str, response_class=ORJSONResponse):
    response = {
        "ok": True,
        "result": {
            "id": 321,
            "is_bot": True,
            "first_name": "SupermemBot",
            "username": "LangAnkBot"
        }
    }
    return ORJSONResponse(response)
