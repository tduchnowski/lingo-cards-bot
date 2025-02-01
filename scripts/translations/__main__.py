import argparse
from translations import Chat, WordList, WordProcessor, ChatError

def translate(path):
    print(path)

def lemmas(path):
    print('lemmas')


def main():
    # cmdline arguments setup
    parser = argparse.ArgumentParser(description="Helper tool for translating or generating word frequency lists")
    subparsers = parser.add_subparsers(dest="command", help="List of commands, [command] -h prints the arguments for the command")
    # translate command
    translate_cmd = subparsers.add_parser('translate', help="translate word lists specified in a .json file")
    translate_cmd.add_argument("config", type=str, help="path to a config file containing word lists locations")
    # lemmas command
    lemmas_cmd = subparsers.add_parser('lemmas', help="transform words into lemmas")
    lemmas_cmd.add_argument('wordlist_path', type=str, help='path to a file containing list of words')
    args = parser.parse_args()
    if args.command == 'translate':
        translate(args.config)
    elif args.command == 'lemmas':
        lemmas(args.wordlist_path)
    else:
        print("No command was given. Nothing to do. Type -h for available commands.")
    # base_dir = os.path.dirname(__file__)
    # words_paths = {
    #     "Polish": os.path.join(base_dir, 'WordLists/Polish.txt'),
    #     "Russian": os.path.join(base_dir, 'WordLists/Russian.txt')
    # }
    # word_lists = [WordList(name, path) for name, path in words_paths.items()]
    # try:
    #     chat = Chat(base_url="https://api.deepseek.com", api_key='', model='deepseek-chat')
    #     for wl in word_lists:
    #         wp = WordProcessor(chat, wl.name, 'English')
    #         #wp.translate_words(wl.words)
    # except ChatError as ce:
    #     print(ce)
    # except Exception as e:
    #     print(e)


if __name__ == "__main__":
    main()

