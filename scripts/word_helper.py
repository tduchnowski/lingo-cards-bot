import argparse
from translations.translator import Chat, WordList, WordProcessor, ChatError
from translations.frequencies import make_lemmas


def translate(path):
    print(path)

def lemmas(path, destination, language, cpus=6):
    lemmas_frequencies = make_lemmas(path, language, cpus=cpus)    
    save_csv(lemmas_frequencies, destination)

def save_csv(frequencies_dict, target_path):
    print(f'Saving to {target_path}')
    header = "word,frequency\n"
    lines = [f"{k}, {v}" for k, v in frequencies_dict.items()]
    with open(target_path, 'w') as f:
        f.write(header)
        f.write('\n'.join(lines))
    print('Done')

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
    lemmas_cmd.add_argument('destination', type=str, help='path of a file where a result is going to be saved')
    lemmas_cmd.add_argument('--language', type=str, help='language of the words that frequency list contains')
    lemmas_cmd.add_argument('--cpus', type=int, help='number of cpus used for generating lemmas')
    args = parser.parse_args()
    if args.command == 'translate':
        translate(args.config)
    elif args.command == 'lemmas':
        lemmas(args.wordlist_path, args.destination, args.language)
    else:
        print("Type -h for available commands.")
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

