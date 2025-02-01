from openai import OpenAI, APIConnectionError


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

    def __repr__(self):
        return f"Chat(base_url={self.base_url}, api_key={self.api_key}, model={self.model})"

    def __call__(self, prompt):
        try:
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "user", "content": prompt},
                ],
                stream=False
            )
            if response.usage is not None: # to avoid LSPs 'prompt_tokens is not a known attribute of None'
                self.total_prompt_tokens += response.usage.prompt_tokens
                self.total_completion_tokens += response.usage.completion_tokens
                return response
        except APIConnectionError:
            raise ChatError("Failed to talk to the chat. Connection Error.")


class WordList:
    def __init__(self, name, path=""):
        self.name = name
        self._path = path
        self.words = []
        self.add_from_file(path)

    def __repr__(self):
        return f"WordList(name=\"{self.name}\", path=\"{self._path}\")"

    def __len__(self):
        if self.words:
            return len(self.words)
        return 0

    def add_from_file(self, path: str):
        """
        Generates a list of words from a .txt file with words separated by a newline
        """
        # TODO better handling of the exception
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
"words":[
    {
        "original": <put original word from the list>
        "meanings":[<put meanings of that word in <target-language>>]
        "examples": [
        {
        "example": <provide an exemple of usage in a sentence in <original-language> for that word>
        "translation":<give translation for provided exemple in <target-language>>
        }
    }
]
}
Give at least 3 examples for each word. In a response give me just JSON.
"""
    PROMPT_TEMPLATE_CORRECTION = """Given a list of <original-language> words: [<word-list>], generate a list of those words in their
root forms. Output just the root forms, each in one line."""

    def __init__(self, chat:Chat, original_language, target_language):
        self.chat = chat
        self.original_language = original_language
        self.target_language = target_language
        self.prompt_template = self.PROMPT_TEMPLATE_TRANSLATION.replace('<target-language>', self.target_language)
        self.prompt_template = self.prompt_template.replace('<original-language>', self.original_language)

    def translate_words(self, words):
        prompt = self.prompt_template.replace('<word-list>', ', '.join(words))
        response = self.chat(prompt)
        return self._convert_response_to_json(response)

    def correct_words(self, words:list) -> list:
        """
        For a given list of words, outputs the root of those words
        For example, "running" should become "to run", "sits" should be "to sit" etc.
        """
        return []

    def _convert_response_to_json(self, response):
        pass
