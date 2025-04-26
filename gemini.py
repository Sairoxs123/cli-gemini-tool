import google.generativeai as genai
import os

def generate_text(prompt):
    """
    Generates text using the Gemini API based on the given prompt.

    Args:
        prompt: The text prompt to use for generating text.

    Returns:
        The generated text, or None if an error occurred.
    """
    try:
        api_key = "AIzaSyCpKacMFM0TalLQA4vCYMNZeuqD-VoDj6E"
        if not api_key:
            raise ValueError("GOOGLE_API_KEY environment variable not set.")
        genai.configure(api_key=api_key)
        model = genai.GenerativeModel('gemini-2.0-flash')
        response = model.generate_content(prompt)
        return response.text
    except Exception as e:
        print(f"An error occurred: {e}")
        return None

if __name__ == "__main__":
    user_prompt = input("Enter your prompt: ")
    generated_text = generate_text(user_prompt)
    if generated_text:
        print("Generated Text:")
        print(generated_text)
