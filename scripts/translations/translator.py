import json
from threading import Lock
from openai import OpenAI, APIConnectionError, RateLimitError


class ChatError(Exception):
    pass


class Chat:
    def __init__(self, base_url, api_key, model):
        self.base_url = base_url
        self.api_key = api_key
        self.model = model
        try:
            self.client = OpenAI(api_key=api_key, base_url=base_url)
        except APIConnectionError:
            raise ChatError(f"Failed to initialize {self}. Connection error.")
        self.total_prompt_tokens = 0
        self.total_completion_tokens = 0
        self.prompt_tokens_mutex = Lock()

    def __repr__(self):
        return f"Chat(base_url={self.base_url}, api_key={self.api_key}, model={self.model})"

    # TODO: do smth about rate limits
    def __call__(self, prompt):
        try:
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "user", "content": prompt},
                ],
                stream=False
            )
            if response.usage is not None: 
                with self.prompt_tokens_mutex:
                    self.total_prompt_tokens += response.usage.prompt_tokens
                    self.total_completion_tokens += response.usage.completion_tokens
                print(f"total tokens used so far: {self.total_completion_tokens + self.total_prompt_tokens}")
                return response.choices[0].message.content
        except APIConnectionError:
            raise ChatError("Failed to talk to the chat. Connection Error.")
        except RateLimitError:
            raise ChatError("Rate limit exceeded")
        except Exception as e:
            print(e)
            raise ChatError("another error for completion")


class WordList:
    def __init__(self, name, path=""):
        self.name = name
        self._path = path
        self.words = []
        self.add_from_csv(path)

    def __repr__(self):
        return f"WordList(name=\"{self.name}\", path=\"{self._path}\")"

    def __len__(self):
        if self.words:
            return len(self.words)
        return 0

    def add_from_csv(self, path: str):
        """
        Loads a list of words from a .txt file with words separated by a newline
        """
        # TODO: better handling of the exception
        try:
            with open(path, 'r') as f:
                self.words += [line.strip() for line in f.readlines()]
        except Exception as e:
            print(e)

    def add(self, *words):
        """
        Add words to self.words
        """
        for w in words:
            self.words.append(w)


class WordProcessor:
    PROMPT_TEMPLATE_TRANSLATION = """Given a list of <original-language> words: [<word-list>], generate a JSON with the following structure:
{
    "words":
    [
        {
            "original": <put original word from the list>,
            "meanings":[<put meanings of that word in English>],
            "lemma": <put a root form of the original word>
            "lemma_meanings":[<put meanings of a lemma in English>],
            "examples": 
            [
                {
                "example": <provide an exemple of usage in a <original-language> sentence for that word>
                "translation":<give an English translation for provided exemple>
                }
            ]
        }
   ]
}
Give at least 3 examples for each word. In a response give me just JSON.
"""
    def __init__(self, chat:Chat, original_language, target_language):
        self.chat = chat
        self.original_language = original_language
        self.target_language = target_language
        self.prompt_template = self.PROMPT_TEMPLATE_TRANSLATION.replace('<target-language>', self.target_language)
        self.prompt_template = self.prompt_template.replace('<original-language>', self.original_language)

    def translate_words(self, words):
        prompt = self.prompt_template.replace('<word-list>', ', '.join(words))
        try:
            response = self.chat(prompt)
        except ChatError as e:
            print(e)
            response = ""
        response_json = self._convert_response_to_json(response)
        words_successful, words_missing = self._summary(words, response_json)
        print(f"batch completed: {len(words_missing)} words missing out of {len(words)}")
        return response_json, words_successful, words_missing

    def _convert_response_to_json(self, response):
        response = response.replace("json\n", "").replace("`","")
        try:
            return json.loads(response)
        except Exception as e:
            print(e)
            return {}

    def _summary(self, words, response_dict):
        words_idxs = {word:i for i, word in enumerate(words)}
        try:
            words_response = [word["original"] for word in response_dict["words"]]
            words_missing = {word:i for word, i in words_idxs.items() if not word in words_response}
            words_successful = {word:i for word, i in words_idxs.items() if word in words_response}
            return words_successful, words_missing
        except Exception as e:
            print(e)
            return {}, words_idxs
