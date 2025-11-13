#!/usr/bin/env/python3

import argparse
import requests
import json
from openai import OpenAI

# Initialising the OpenAI client
client = OpenAI(llm_model="gpt-4o-mini")

app = Flask(__name__)

# def rank_user_submission(user_submission: str) -> {str: str, str: int}:

# Main function to receive the user submission
@app.route("/webhook", methods=["POST"])
def main() {
    data = request.json
    print("Received data: ", data)

    return jsonify({"message": "Data received successfully"}), 200
}


if __name__ == "__main__":
    main();