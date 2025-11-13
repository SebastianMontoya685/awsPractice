#!/usr/bin/env/python3

import argparse
import requests
import json
from openai import OpenAI
from flask import Flask, request, jsonify

# Initialising the OpenAI client
# client = OpenAI(llm_model="gpt-4o-mini")

app = Flask(__name__)
client = OpenAI()

# Function to rank the user submission based on user sentiment
def rank_user_submission(user_submission: str) -> {str: str, str: int}:
    # Creating the prompt for the LLM
    prompt = f"Rank the user submission based on user sentiment, so based on how the user feels: {user_submission}"

    response = client.chat.completions.create(
        # Using the GPT-4o-mini model
        model="gpt-4o-mini",
        messages=[
            {"role": "system", "content": "You are a helpful assistant that ranks user submissions based on user sentiment."},
            {"role": "user", "content": prompt}
        ]
    )

    return response.choices[0].message.content.strip()

# Main function to receive the user submission
@app.route("/webhook", methods=["POST"])
def webhook():
    # Getting the data from the request
    data = request.json.get("data", {})
    print("Received data: ", data)

    # Feeding the data into the LLM to rank the user submission based on user sentiment:
    sentiment = rank_user_submission(data)
    print("Sentiment: ", sentiment)

    # Checking if the sentiment is None
    if sentiment is None:
        return jsonify({"message": "Error ranking user submission"}), 500

    return jsonify({"message": "User submission ranked successfully", "sentiment": sentiment}), 200


if __name__ == "__main__":
    app.run(port=5000)