from translations.frequencies import is_word_invalid

def test_is_word_invalid():
    too_short = ["a", "ab", "abc", "a", "ab", "aa"]
    assert all(is_word_invalid(w) for w in too_short) == True
    contain_digit = ["1234", "aa5ndnd", "112aaa1"]
    assert all(is_word_invalid(w) for w in contain_digit) == True
    contain_forbidden_char = ["abc!!", "\"abcdbsb\"", "**abcb"]
    assert all(is_word_invalid(w) for w in contain_forbidden_char) == True
    valid_words = ["abcd", "ąćśćć"]
    assert any(is_word_invalid(w) for w in valid_words) == False
