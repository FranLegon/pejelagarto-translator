You are an expert Golang developer. Your task is to create a complete, locally runnable web application that functions as a translator for a fictional language called "Pejelagarto". The project will involve a Go backend and a dynamic frontend powered by HTMX and simple, vanilla JavaScript (no npm or external libraries allowed).

Core Task: Create a Go application in a single main.go file that starts a local web server. The server will provide a simple, interactive UI to translate text to and from "Pejelagarto". The translation logic must be perfectly reversible.

Backend Requirements (Go):

applyMapReplacementsToPejelagarto(input string):
    Translation Maps: The core logic relies on three distinct replacement maps (map[string]string).
    wordMap: Replaces words or syllables (e.g., {"hello": "jetzo", "world": "vorlag"}).
    conjunctionMap: Replaces letter pairs (e.g., {"ch": "jj", "sh": "xx"}).
    letterMap: Replaces single, invertible letters (e.g., {"a": "i", "i": "a"}).
    For all maps the condition len(key)<=len(value) must be met.
    Core Replacement Logic:
    All three replacement maps must be unified into a single indexed map (map[int32]map[string]string), making sure index is ordered by len(key) desc, alphabetical(key) asc.
    All multi-rune values (not keys) will be prefixed with a single quote ('). (e.g {25:{"ch": "jj"}} becomes {25:{"ch": "'jj"}})
    To make the map bijective, the inverse of the unified map (-index:{value:key}) (map[int32]map[string]string) will be appended to the unified map.
    Replacements will then be applied using bijectiveMap in order by index desc.
    To prevent re-replacing, every successful replacement must be wrapped in temporary tags (<rep> and </rep>). These tags are removed in a final step before returning.
    Respect casing by casing the same letter positions in original text. For exaple, if a 5 letter word is replaced by a 6 letter word and 2nd and 4th letters are upper case, 2nd and 4th letters will be uppercase in result too.
applyMapReplacementsFromPejelagarto(input string):
    does exactly the same as applyMapReplacementsToPejelagarto but replacements must be applied using bijectiveMap in order by index asc.

applyNumbersLogicToPejelagarto(input string):
    Converts all numbers in string from base 10 to base 7. doesnt remove front-trailing zeros (example: 007 becomes 0010, 00 becomes 00).
applyNumbersLogicFromPejelagarto(input string):
    Converts all numbers in string from base 7 to base 10. doesnt remove front-trailing zeros (example: 0011 becomes 008, 000 becomes 000).


applyAccentReplacementLogicToPejelagarto(input string):
    This funciton will take the count of runes in input and factorize it into primes. Then, for each factor n, change the accent of the nth vowel. (y is always considered a vowel)

    There will be two accent "wheels" (ordered slices), oneRuneAccentsWheel and twoRunesAccentsWheel. The accent at factor n position will be replaced by an accent (from its corresponfing wheel based on ammount of runes) indexed above current accent index. the ammount of places to move forward is determined by the power of the prime factor.

    var oneRuneAccentsWheel = []map[rune]string{
        // No accent (base vowels)
        {'a': "a", 'e': "e", 'i': "i", 'o': "o", 'u': "u", 'y': "y"},
        // Grave accent
        {'a': "à", 'e': "è", 'i': "ì", 'o': "ò", 'u': "ù", 'y': "ỳ"},
        // Acute accent
        {'a': "á", 'e': "é", 'i': "í", 'o': "ó", 'u': "ú", 'y': "ý"},
        // Circumflex accent
        {'a': "â", 'e': "ê", 'i': "î", 'o': "ô", 'u': "û", 'y': "ŷ"},
        // Tilde accent
        {'a': "ã", 'e': "ẽ", 'i': "ĩ", 'o': "õ", 'u': "ũ", 'y': "ỹ"},
        // Ring above
        {'a': "å", 'e': "e̊", 'i': "i̊", 'o': "o̊", 'u': "ů", 'y': "ẙ"},
        // Diaeresis
        {'a': "ä", 'e': "ë", 'i': "ï", 'o': "ö", 'u': "ü", 'y': "ÿ"},
        // Macron
        {'a': "ā", 'e': "ē", 'i': "ī", 'o': "ō", 'u': "ū", 'y': "ȳ"},
        // Breve
        {'a': "ă", 'e': "ĕ", 'i': "ĭ", 'o': "ŏ", 'u': "ŭ", 'y': "y̆"},
    }

    For example, if total count is 245, it factorizes into 5*7*7. So, the fifth vowel will have its accent moved to one further from original in the accent wheel and seventh vowel will have its accent moved two places forward. We call it a wheel because it should loop. Changing accents shouldn't change the totalRuneCount.

applyAccentReplacementLogicFromPejelagarto(input string):
    will just do the same as applyMapReplacementsToPejelagarto but move the accents backwards in the wheels instead of forwards

applyCaseReplacementLogic(input string):
    Lets add CaseReplacementLogic. This will be similar to AccentLogic, but it will use wordCount instead of characterCount. If wordCount is odd, all nth character where n is in the Fibonacci sequence will have its capitalization inverted (if lowercase, made uppercase, if uppercase, made lowercase). If wordCount is even, use Tribonacci sequence instead. Should not change word count.

removeISO8601timestamp(input string):
    if an ISO 8601 timestamp is found in the last line (using regex), remove the line.

addISO8601timestamp(input string, timestamp string):
    adds timestamp as new last string to input

addEmojiDatetimeEncoding(input string):
    Using 5 maps dayEmojiIndex (indexes 0 to 30), monthEmojiIndex (indexes 0 to 11), yearEmojiIndex (indexes 0 to 98), hourEmojiIndex (indexes 0 to 23), minuteEmojiIndex(indexes 0 to 59) and current UTC datetime, choose corresponding emoji from each map and insert them into input string in random places next to a space or a linebreak.

removeAllEmojies(input strings):
    removes emojies

readTimestampUsingEmojiEncoding(input string):
    Using the same 5 maps, locate all emojis and returns ISO 8601 timestamp as string. Remove the emojies.

applyPuntuationReplacementsToPejelagarto(input string):
    just like applyMapReplacementsToPejelagarto, but uses it's own map[int32]map[string] punctuationReplacementsMap. Must also be made bijective.
    E.g: {1:{"?","‽",}{2:{".",".."},{3:{"'","〝"}}}}

applyPuntuationReplacementsFromPejelagarto(input string):
    just like just like applyMapReplacementsToPejelagarto but uses punctuationReplacementsMap.





Main Translation Functions:

func TranslateToPejelagarto(input string) string {
    input = removeAllEmojies(input)
    input = removeISO8601timestamp(input)
	input = applyNumbersFromBase10ToBase7(input)
    input = applyPuntuationReplacementsToPejelagarto(input)
	input = applyMapReplacementsToPejelagarto(input)
	input = applyAccentReplacementLogicToPejelagarto(input)
	input = applyCaseReplacementLogicToPejelagarto(input)
    input = addEmojiDatetimeEncoding(input string)
    return input
}
// 	translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
    timestamp := readTimestampUsingEmojiEncoding(input)
    input = removeAllEmojies(input)
	input = applyCaseReplacementLogicFromPejelagarto(input)
	input = applyAccentReplacementLogicFromPejelagarto(input)
	input = applyMapReplacementsFromPejelagarto(input)
    input = applyPuntuationReplacementsFromPejelagarto(input)
	input = applyNumbersFromBase7ToBase10(input)
    input = addISO8601timestamp(input, timestamp)
    return input
}

Web Server:
The main function should start a web server on http://localhost:8080.
It must serve the main HTML UI from the root path (/).
It must handle POST requests from the frontend at /to and /from, using the translation functions to process the request body and return the translated text.
Frontend Requirements (HTML/JS/HTMX):

Single File UI: The complete HTML, CSS, and JavaScript must be stored in a Go string constant and served by the web server.
UI Elements:
A main title: "Pejelagarto Translator".
An input textarea with a label (e.g., "Human").
An output textarea (read-only) with a label (e.g., "Pejelagarto").
A single main translation button (e.g., "Translate to Pejelagarto").
An "Invert" button (e.g., with "⇅" or similar icon).
A checkbox with the label "Live Translation".
Interactivity (JavaScript & HTMX):
Invert Button: When clicked, the JavaScript should:
Swap the text content of the input and output textarea elements.
Swap the labels for the textarea elements.
Toggle the state of the main translation button: change its text (e.g., "Translate to Pejelagarto" <-> "Translate from Pejelagarto") and update its hx-post attribute to point to the correct endpoint (/to or /from).
Live Translation Checkbox: When checked, the JavaScript should:
Hide the main translation button.
Add an event listener to the input textarea (e.g., on input or keyup).
On input, automatically trigger the HTMX request that the (now hidden) button would have, to perform the translation in real-time.
When unchecked, the JavaScript should show the main translation button again and remove the live-translation event listener.
HTMX Core: The main translation button (and the live translation) will use HTMX attributes (hx-post, hx-trigger, hx-target, hx-include) to send the input textarea content to the backend and place the response in the output textarea.
Testing Requirements:

Reversibility Test: Create a main_test.go file.
Inside it, create a function TestReversibility that verifies the translation is perfectly reversible using fuzzy testing module with 100000 random seeds. It should only stop if it finds a test fail and display which tests passed and which failed, showing value of input string before and after each function.
The tests are:
applyNumbersFromBase10ToBase7(applyNumbersFromBase7ToBase10(input)) == input
applyNumbersFromBase7ToBase10(applyNumbersFromBase10ToBase7(input)) == input
applyMapReplacementsToPejelagarto(applyMapReplacementsFromPejelagarto(input)) = input
applyMapReplacementsFromPejelagarto(applyMapReplacementsToPejelagarto(input)) = input
applyAccentReplacementLogicToPejelagarto(applyAccentReplacementLogicFromPejelagarto(input)) == input
applyAccentReplacementLogicFromPejelagarto(applyAccentReplacementLogicToPejelagarto(input)) == input
applyCaseReplacementLogicToPejelagarto(applyCaseReplacementLogicFromPejelagarto(input)) == input
applyCaseReplacementLogicFromPejelagarto(applyCaseReplacementLogicToPejelagarto(input)) == input
TranslateToPejelagarto(TranslateFromPejelagarto(input)) == input
TranslateFromPejelagarto(TranslateToPejelagarto(input)) == input


Final Deliverable: A single, well-commented main.go file and a main_test.go file. The comments should include the build command needed to compile the server.

Definition of done: all tests must pass multiple rounds of fuzzy testing many random seeds for the project to be considered done.


