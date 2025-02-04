import argparse
from translations.frequencies import extract_freqs, process_batch


def translate(path):
    print(path)

def validate(path, destination):
    frequencies = extract_freqs(path)
    filtered = process_batch(frequencies)    
    save_csv(filtered, destination)

def save_csv(word_freqs, target_path):
    print(f'Saving to {target_path}')
    header = "word,frequency\n"
    lines = [f"{wf.word}, {wf.freq}" for wf in word_freqs]
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
    lemmas_cmd = subparsers.add_parser('validate', help="transform words into lemmas")
    lemmas_cmd.add_argument('wordlist_path', type=str, help='path to a file containing list of words')
    lemmas_cmd.add_argument('destination', type=str, help='path of a file where a result is going to be saved')
    lemmas_cmd.add_argument('--cpus', type=int, help='number of cpus used for generating lemmas')
    args = parser.parse_args()
    if args.command == 'translate':
        translate(args.config)
    elif args.command == 'validate':
        validate(args.wordlist_path, args.destination)
    else:
        print("Type -h for available commands.")


if __name__ == "__main__":
    main()
