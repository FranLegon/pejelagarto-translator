//go:build ignore
// +build ignore

// Simple HTTP server for frontend mode
// This serves the WASM-enabled UI and TTS endpoints only
// Translation happens client-side in the browser via WASM

package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

	"pejelagarto-translator/obfuscation"
)

//go:embed get-requirements.ps1 get-requirements.sh
var embeddedGetRequirements embed.FS

// Downloadable feature - empty by default, can be built with embedded binaries
var embeddedBinaries embed.FS

const isDownloadable = false

var pronunciationLanguage string
var pronunciationLanguageDropdown bool
var tempRequirementsDir string

// Audio cache for storing normal and slow versions
var audioCache = struct {
	sync.RWMutex
	cache map[string][]byte // key: text+lang hash -> audio data
}{
	cache: make(map[string][]byte),
}

// Frontend HTML - identical UI to regular mode, but uses WASM for translation
// This constant is based on htmlUI from main.go with only the handleLiveTranslation() function modified
const htmlUIFrontend = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pejelagarto Translator</title>
    <script src="/wasm_exec.js"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --bg-gradient-start: #1a1a2e;
            --bg-gradient-end: #16213e;
            --container-bg: #0f3460;
            --text-primary: #e1e1e1;
            --text-secondary: #b0b0b0;
            --heading-color: #53a8e2;
            --button-gradient-start: #53a8e2;
            --button-gradient-end: #3d7ea6;
            --button-shadow: rgba(83, 168, 226, 0.4);
            --button-hover-shadow: rgba(83, 168, 226, 0.6);
            --invert-btn-gradient-start: #e94560;
            --invert-btn-gradient-end: #d62839;
            --invert-btn-shadow: rgba(233, 69, 96, 0.4);
            --invert-btn-hover-shadow: rgba(233, 69, 96, 0.6);
            --border-color: #2a2a40;
            --textarea-bg: #1a1a2e;
            --textarea-readonly-bg: #16213e;
            --textarea-focus-border: #53a8e2;
            --theme-btn-bg: #53a8e2;
            --theme-btn-hover: #3d7ea6;
        }

        [data-theme="light"] {
            --bg-gradient-start: #667eea;
            --bg-gradient-end: #764ba2;
            --container-bg: white;
            --text-primary: #333;
            --text-secondary: #666;
            --heading-color: #667eea;
            --button-gradient-start: #667eea;
            --button-gradient-end: #764ba2;
            --button-shadow: rgba(102, 126, 234, 0.4);
            --button-hover-shadow: rgba(102, 126, 234, 0.6);
            --invert-btn-gradient-start: #f093fb;
            --invert-btn-gradient-end: #f5576c;
            --invert-btn-shadow: rgba(245, 87, 108, 0.4);
            --invert-btn-hover-shadow: rgba(245, 87, 108, 0.6);
            --border-color: #e0e0e0;
            --textarea-bg: white;
            --textarea-readonly-bg: #f5f5f5;
            --textarea-focus-border: #667eea;
            --theme-btn-bg: #ffd700;
            --theme-btn-hover: #ffed4e;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
            transition: background 0.3s ease;
        }
        
        .container {
            background: var(--container-bg);
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 900px;
            width: 100%;
            position: relative;
            transition: background 0.3s ease;
        }
        
        .theme-toggle {
            position: absolute;
            top: 20px;
            right: 20px;
            background: var(--theme-btn-bg);
            border: none;
            border-radius: 50%;
            width: 45px;
            height: 45px;
            cursor: pointer;
            display: flex;
            justify-content: center;
            align-items: center;
            font-size: 24px;
            transition: all 0.3s ease;
            box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
            z-index: 10;
        }
        
        .theme-toggle:hover {
            background: var(--theme-btn-hover);
            transform: scale(1.1) rotate(15deg);
            box-shadow: 0 6px 15px rgba(0, 0, 0, 0.3);
        }
        
        h1 {
            text-align: center;
            color: var(--heading-color);
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
            transition: color 0.3s ease;
        }
        
        .translator-box {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .text-area-container {
            display: flex;
            flex-direction: column;
        }
        
        label {
            font-weight: bold;
            margin-bottom: 8px;
            color: var(--text-primary);
            font-size: 1.1em;
            transition: color 0.3s ease;
            display: flex;
            align-items: center;
            gap: 8px;
            min-height: 40px;
        }
        
        textarea {
            width: 100%;
            height: 250px;
            padding: 15px;
            border: 2px solid var(--border-color);
            border-radius: 10px;
            font-size: 14px;
            font-family: 'Courier New', monospace;
            resize: none;
            transition: all 0.3s ease;
            background-color: var(--textarea-bg);
            color: var(--text-primary);
        }
        
        textarea:focus {
            outline: none;
            border-color: var(--textarea-focus-border);
        }
        
        textarea[readonly] {
            background-color: var(--textarea-readonly-bg);
            cursor: not-allowed;
            resize: none;
        }
        
        .controls {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        
        button {
            background: linear-gradient(135deg, var(--button-gradient-start) 0%, var(--button-gradient-end) 100%);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: background 0.3s ease;
            box-shadow: 0 4px 15px var(--button-shadow);
        }
        
        button:active {
            transform: translateY(0);
        }
        
        .invert-btn {
            background: linear-gradient(135deg, var(--invert-btn-gradient-start) 0%, var(--invert-btn-gradient-end) 100%);
            padding: 12px 20px;
            font-size: 20px;
            box-shadow: 0 4px 15px var(--invert-btn-shadow);
        }
        
        .checkbox-container {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 16px;
            color: var(--text-primary);
            transition: color 0.3s ease;
        }
        
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }
        
        .play-btn {
            background: linear-gradient(135deg, #56ab2f 0%, #a8e063 100%);
            padding: 8px 16px;
            font-size: 18px;
            box-shadow: 0 4px 15px rgba(86, 171, 47, 0.4);
            min-width: auto;
        }
        
        .play-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        
        .hidden {
            display: none !important;
        }
        
        @media (max-width: 768px) {
            body {
                padding: 10px;
                align-items: flex-start;
            }
            
            .container {
                padding: 15px;
                border-radius: 15px;
                margin-top: 10px;
            }
            
            .translator-box {
                grid-template-columns: 1fr;
                gap: 15px;
                margin-bottom: 15px;
            }
            
            h1 {
                font-size: 1.5em;
                margin-bottom: 20px;
                padding-right: 50px;
            }
            
            .theme-toggle {
                width: 40px;
                height: 40px;
                top: 15px;
                right: 15px;
                font-size: 20px;
            }
            
            label {
                font-size: 0.95em;
                margin-bottom: 6px;
            }
            
            textarea {
                height: 120px;
                padding: 10px;
                font-size: 13px;
            }
            
            button {
                padding: 10px 20px;
                font-size: 14px;
            }
            
            .play-btn {
                padding: 6px 12px;
                font-size: 16px;
            }
            
            .invert-btn {
                padding: 10px 16px;
                font-size: 18px;
            }
            
            .controls {
                gap: 10px;
            }
            
            .checkbox-container {
                font-size: 14px;
            }
            
            input[type="checkbox"] {
                width: 18px;
                height: 18px;
            }
        }
        
        .htmx-indicator {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid var(--border-color);
            border-top: 3px solid var(--button-gradient-start);
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-left: 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .download-section {
            margin-top: 30px;
            padding: 15px;
            background: var(--textarea-bg);
            border-radius: 8px;
            border: 1px solid var(--border-color);
        }
        
        .download-buttons {
            display: flex;
            gap: 10px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .download-btn {
            display: inline-block;
            padding: 8px 16px;
            background: linear-gradient(135deg, var(--button-gradient-start), var(--button-gradient-end));
            color: var(--text-primary);
            text-decoration: none;
            border-radius: 6px;
            font-size: 13px;
            font-weight: 500;
            transition: all 0.3s ease;
            border: 1px solid var(--border-color);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
        }
        
        .download-btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.25);
            filter: brightness(1.1);
        }
        
        .download-btn:active {
            transform: translateY(0);
            box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
        }
        
        /* Desktop: align download section to bottom left */
        @media (min-width: 769px) {
            .download-section {
                position: fixed;
                bottom: 20px;
                left: 20px;
                margin-top: 0;
                max-width: 280px;
                z-index: 100;
            }
            
            .download-buttons {
                flex-direction: column;
                gap: 8px;
            }
            
            .download-btn {
                width: 100%;
                text-align: center;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <button class="theme-toggle" onclick="toggleTheme()" aria-label="Toggle theme">
            <span id="theme-icon">🌙</span>
        </button>
        <h1>🐊 Pejelagarto Translator 🐊</h1>
        
        <div class="translator-box">
            <div class="text-area-container">
                <label id="input-label">Human:</label>
                <textarea id="input-text" placeholder="Type your text here..."></textarea>
            </div>
            
            <div class="text-area-container">
                <label id="output-label">Pejelagarto: <button class="play-btn" id="play-output" onclick="playAudio('output', false)">🔊 Play</button>{{DROPDOWN_PLACEHOLDER}}</label>
                <textarea id="output-text" readonly placeholder="Translation will appear here..."></textarea>
            </div>
        </div>
        
        <div class="controls">
            <button 
                id="translate-btn"
                onclick="handleTranslateClick()">
                Translate to Pejelagarto
            </button>
            
            <button class="invert-btn" onclick="invertTranslation()">⇅</button>
            
            <div class="checkbox-container">
                <input type="checkbox" id="live-translate" onchange="toggleLiveTranslation()" checked>
                <label for="live-translate" style="margin: 0;">Live Translation</label>
            </div>
            
            <span id="loading-indicator" class="htmx-indicator"></span>
        </div>
    </div>
    
    <div id="download-section" class="download-section" style="display: none;">
        <h3 style="color: var(--text-primary); margin-bottom: 10px; font-size: 16px;">Download Translator</h3>
        <div class="download-buttons">
            <a href="/download/windows" download="pejelagarto-translator.exe" class="download-btn">
                💻 Windows
            </a>
            <a href="/download/linux" download="pejelagarto-translator" class="download-btn">
                🐧 Linux/Mac
            </a>
        </div>
    </div>
    
    <script>
        let isInverted = false;
        let liveTranslateEnabled = true;
        let wasmReady = false;
        
        // Load WASM module
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/translator.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
            wasmReady = true;
            console.log('WASM module loaded and ready');
        }).catch((err) => {
            console.error("Failed to load WASM:", err);
            alert('Failed to load translation module. Please refresh the page.');
        });
        
        // Initialize theme on page load
        (function initTheme() {
            const savedTheme = localStorage.getItem('theme') || 'dark';
            document.documentElement.setAttribute('data-theme', savedTheme);
            updateThemeIcon(savedTheme);
        })();
        
        // Initialize live translation on page load
        (function initLiveTranslation() {
            const translateBtn = document.getElementById('translate-btn');
            const inputText = document.getElementById('input-text');
            
            // Hide translate button since live translation is on
            translateBtn.classList.add('hidden');
            
            // Add event listener for live translation
            inputText.addEventListener('input', handleLiveTranslation);
        })();
        
        // Check if downloadable version and show download section
        (function initDownloadSection() {
            fetch('/api/is-downloadable')
                .then(response => response.json())
                .then(data => {
                    if (data.downloadable) {
                        document.getElementById('download-section').style.display = 'block';
                    }
                })
                .catch(err => console.log('Download check failed:', err));
        })();
        
        function toggleTheme() {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            updateThemeIcon(newTheme);
        }
        
        function updateThemeIcon(theme) {
            const icon = document.getElementById('theme-icon');
            icon.textContent = theme === 'dark' ? '🌙' : '☀️';
        }
        
        function handleTranslateClick() {
            handleLiveTranslation();
        }
        
        function invertTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const translateBtn = document.getElementById('translate-btn');
            
            const temp = inputText.value;
            inputText.value = outputText.value;
            outputText.value = temp;
            
            isInverted = !isInverted;
            
            if (isInverted) {
                translateBtn.textContent = 'Translate from Pejelagarto';
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
            }
            
            resetToSingleButton();
        }
        
        function toggleLiveTranslation() {
            const checkbox = document.getElementById('live-translate');
            const translateBtn = document.getElementById('translate-btn');
            const inputText = document.getElementById('input-text');
            
            liveTranslateEnabled = checkbox.checked;
            
            if (liveTranslateEnabled) {
                translateBtn.classList.add('hidden');
                inputText.addEventListener('input', handleLiveTranslation);
                handleLiveTranslation();
            } else {
                translateBtn.classList.remove('hidden');
                inputText.removeEventListener('input', handleLiveTranslation);
            }
        }
        
        // MODIFIED FOR WASM: Use WASM functions instead of fetch
        function handleLiveTranslation() {
            if (!wasmReady) {
                console.log('WASM not ready yet, skipping translation');
                return;
            }
            
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            
            try {
                if (!isInverted) {
                    // Human to Pejelagarto
                    outputText.value = GoTranslateToPejelagarto(inputText.value);
                } else {
                    // Pejelagarto to Human
                    outputText.value = GoTranslateFromPejelagarto(inputText.value);
                }
            } catch (error) {
                console.error('Translation error:', error);
            }
        }
        
        // Rest of the JavaScript is identical to main.go (TTS functions, etc.)
        let currentOutputText = '';
        let currentLanguage = '';
        let currentInvertedState = false;
        let slowAudioReady = {};
        
        function watchOutputChanges() {
            const outputText = document.getElementById('output-text');
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            if (outputText.value !== currentOutputText || selectedLanguage !== currentLanguage || isInverted !== currentInvertedState) {
                currentOutputText = outputText.value;
                currentLanguage = selectedLanguage;
                currentInvertedState = isInverted;
                resetToSingleButton();
                
                const cacheKey = currentOutputText + ':' + selectedLanguage;
                if (slowAudioReady[cacheKey]) {
                    const source = isInverted ? 'input' : 'output';
                    const container = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
                    splitButton(source, container);
                }
            }
        }
        
        function resetToSingleButton() {
            const outputLabel = document.getElementById('output-label');
            const inputLabel = document.getElementById('input-label');
            
            const oldDropdown = document.getElementById('tts-language');
            const selectedLang = oldDropdown ? oldDropdown.value : 'russian';
            
            const dropdownHTML = document.getElementById('tts-language') ? 
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="kazakh">North-North-East</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">South-East-East</option><option value="hindi">South-East</option><option value="icelandic">South-South-East</option><option value="romanian">South</option><option value="vietnamese">South-South-West</option><option value="swahili">South-West</option><option value="swedish">South-West-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option></select>' : '';
            
            if (isInverted) {
                inputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-input" onclick="playAudio(&quot;input&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                outputLabel.textContent = 'Human:';
            } else {
                outputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-output" onclick="playAudio(&quot;output&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                inputLabel.textContent = 'Human:';
            }
            
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
        
        setInterval(watchOutputChanges, 500);
        
        function playAudio(source, slow) {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const playInputBtn = document.getElementById('play-input');
            const playOutputBtn = document.getElementById('play-output');
            const playInputSlowBtn = document.getElementById('play-input-slow');
            const playOutputSlowBtn = document.getElementById('play-output-slow');
            
            let textToSpeak = '';
            let button = null;
            let container = null;
            
            if (source === 'input') {
                textToSpeak = inputText.value;
                button = slow ? playInputSlowBtn : playInputBtn;
                container = document.getElementById('input-label');
            } else {
                textToSpeak = outputText.value;
                button = slow ? playOutputSlowBtn : playOutputBtn;
                container = document.getElementById('output-label');
            }
            
            if (!textToSpeak || textToSpeak.trim() === '') {
                alert('No text to convert to speech!');
                return;
            }
            
            button.disabled = true;
            const originalText = button.textContent;
            button.textContent = '⏳';
            
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            let url = selectedLanguage ? '/tts?lang=' + selectedLanguage : '/tts';
            if (slow) {
                url += (selectedLanguage ? '&' : '?') + 'slow=true';
            }
            
            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain'
                },
                body: textToSpeak
            })
            .then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText || 'TTS request failed: ' + response.statusText);
                }
                return response.blob();
            })
            .then(blob => {
                const audioUrl = URL.createObjectURL(blob);
                const audio = new Audio(audioUrl);
                
                audio.onended = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                };
                
                audio.onerror = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                    alert('Error playing audio');
                };
                
                audio.play();
                button.textContent = '▶️';
                
                if (!slow) {
                    checkForSlowAudio(textToSpeak, selectedLanguage, source, container);
                }
            })
            .catch(error => {
                console.error('TTS error:', error);
                button.disabled = false;
                button.textContent = originalText;
                
                let errorMsg = error.message;
                if (errorMsg.includes('voice model not found')) {
                    const lang = selectedLanguage || 'portuguese';
                    errorMsg = 'Language model not installed for: ' + lang + '\\n\\nTo install the model, run:\\ncd tts/requirements/piper/languages\\n.\\\\download_models.ps1\\n\\nOr download manually from:\\ntts/requirements/piper/languages/README.md';
                }
                
                alert('Text-to-speech error:\\n\\n' + errorMsg);
            });
        }
        
        function checkForSlowAudio(text, language, source, container) {
            const url = language ? '/tts-check-slow?lang=' + language : '/tts-check-slow';
            const cacheKey = text + ':' + language;
            
            const checkInterval = setInterval(() => {
                fetch(url, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'text/plain'
                    },
                    body: text
                })
                .then(response => response.json())
                .then(data => {
                    if (data.ready) {
                        clearInterval(checkInterval);
                        slowAudioReady[cacheKey] = true;
                        splitButton(source, container);
                    }
                })
                .catch(error => {
                    console.error('Error checking slow audio:', error);
                    clearInterval(checkInterval);
                });
            }, 1000);
            
            setTimeout(() => clearInterval(checkInterval), 30000);
        }
        
        function splitButton(source, container) {
            const expectedSource = isInverted ? 'input' : 'output';
            if (source !== expectedSource) {
                return;
            }
            
            const expectedContainer = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
            if (container !== expectedContainer) {
                return;
            }
            
            const oldDropdown = document.getElementById('tts-language');
            const selectedLang = oldDropdown ? oldDropdown.value : 'russian';
            
            const dropdownHTML = document.getElementById('tts-language') ? 
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="kazakh">North-North-East</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">South-East-East</option><option value="hindi">South-East</option><option value="icelandic">South-South-East</option><option value="romanian">South</option><option value="vietnamese">South-South-West</option><option value="swahili">South-West</option><option value="swedish">South-West-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option></select>' : '';
            
            const label = 'Pejelagarto:';
            const buttonId = source === 'input' ? 'play-input' : 'play-output';
            const slowButtonId = source === 'input' ? 'play-input-slow' : 'play-output-slow';
            
            container.innerHTML = label + 
                ' <button class="play-btn" id="' + buttonId + '" onclick="playAudio(&quot;' + source + '&quot;, false)" style="width: 50px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🐇🔊</button>' +
                ' <button class="play-btn" id="' + slowButtonId + '" onclick="playAudio(&quot;' + source + '&quot;, true)" style="width: 50px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🐌🔊</button>' +
                dropdownHTML;
            
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
    </script>
</body>
</html>`

// Handler for serving the frontend UI
func handleFrontendIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := htmlUIFrontend
	if pronunciationLanguageDropdown {
		dropdownHTML := ` <select id="tts-language" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;">
                    <option value="russian">North</option>
                    <option value="kazakh">North-North-East</option>
                    <option value="german">North-East</option>
                    <option value="turkish">North-East-East</option>
                    <option value="portuguese">East</option>
                    <option value="french">South-East-East</option>
                    <option value="hindi">South-East</option>
                    <option value="icelandic">South-South-East</option>
                    <option value="romanian">South</option>
                    <option value="vietnamese">South-South-West</option>
                    <option value="swahili">South-West</option>
                    <option value="swedish">South-West-West</option>
                    <option value="czech">West</option>
                    <option value="chinese">North-West-West</option>
                    <option value="norwegian">North-West</option>
                    <option value="hungarian">North-North-West</option>
                </select>`
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", dropdownHTML, 1)
	} else {
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", "", 1)
	}

	fmt.Fprint(w, html)
}

// handleIsDownloadable returns JSON indicating if this build supports downloads
func handleIsDownloadable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if isDownloadable {
		fmt.Fprint(w, `{"downloadable": true}`)
	} else {
		fmt.Fprint(w, `{"downloadable": false}`)
	}
}

// handleDownloadWindows serves the embedded Windows binary
func handleDownloadWindows(w http.ResponseWriter, r *http.Request) {
	if !isDownloadable {
		http.Error(w, "Downloads not available in this build", http.StatusNotFound)
		return
	}

	data, err := embeddedBinaries.ReadFile("bin/pejelagarto-translator.exe")
	if err != nil {
		http.Error(w, "Windows binary not found", http.StatusNotFound)
		if !obfuscation.Obfuscated() {
			log.Printf("Error reading Windows binary: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=pejelagarto-translator.exe")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

// handleDownloadLinux serves the embedded Linux/Mac binary
func handleDownloadLinux(w http.ResponseWriter, r *http.Request) {
	if !isDownloadable {
		http.Error(w, "Downloads not available in this build", http.StatusNotFound)
		return
	}

	data, err := embeddedBinaries.ReadFile("bin/pejelagarto-translator")
	if err != nil {
		http.Error(w, "Linux/Mac binary not found", http.StatusNotFound)
		if !obfuscation.Obfuscated() {
			log.Printf("Error reading Linux/Mac binary: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=pejelagarto-translator")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

func main() {
	// Parse flags
	pronunciationLangFlag := flag.String("pronunciation_language", "russian", "TTS pronunciation language")
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, "Show language dropdown in UI for TTS")
	flag.Parse()

	pronunciationLanguage = *pronunciationLangFlag
	pronunciationLanguageDropdown = *pronunciationLangDropdownFlag

	if !obfuscation.Obfuscated() {
		log.Println("Starting Pejelagarto Translator server")
		log.Println("Translation: Client-side (WebAssembly)")
		log.Println("TTS Audio: Server-side")
		log.Printf("TTS Language: %s\n", pronunciationLanguage)
	}

	// Initialize TTS
	if !obfuscation.Obfuscated() {
		log.Println("Initializing TTS requirements...")
	}
	var languageToDownload string
	if !*pronunciationLangDropdownFlag {
		// Dropdown is disabled, download only the selected language
		languageToDownload = *pronunciationLangFlag
	}
	// If dropdown is enabled, languageToDownload remains empty and all languages are downloaded
	if err := extractEmbeddedRequirements(languageToDownload); err != nil {
		log.Fatalf("Failed to extract TTS requirements: %v", err)
	}

	// Serve static files
	http.HandleFunc("/", handleFrontendIndex)
	http.HandleFunc("/translator.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, "bin/translator.wasm")
	})
	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "bin/wasm_exec.js")
	})

	// TTS endpoints
	http.HandleFunc("/tts", handleTextToSpeech)
	http.HandleFunc("/tts-check-slow", handleCheckSlowAudio)

	// Download endpoints
	http.HandleFunc("/api/is-downloadable", handleIsDownloadable)
	http.HandleFunc("/download/windows", handleDownloadWindows)
	http.HandleFunc("/download/linux", handleDownloadLinux)

	addr := ":8080"
	url := "http://localhost:8080"

	if !obfuscation.Obfuscated() {
		log.Printf("Server starting on %s\n", url)
	}

	// Open browser
	time.Sleep(500 * time.Millisecond)
	openBrowser(url)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// TTS Functions (server_frontend.go is standalone due to //go:build ignore)

func extractEmbeddedRequirements(singleLanguage string) error {
	// Determine the requirements directory based on OS
	var requirementsDir string
	if runtime.GOOS == "windows" {
		requirementsDir = filepath.Join(os.TempDir(), "pejelagarto-translator", "requirements")
	} else {
		requirementsDir = filepath.Join("/tmp", "pejelagarto-translator", "requirements")
	}

	tempRequirementsDir = requirementsDir

	// Check what dependencies are missing
	piperExe := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		piperExe += ".exe"
	}

	espeakData := filepath.Join(tempRequirementsDir, "espeak-ng-data")
	piperDir := filepath.Join(tempRequirementsDir, "piper")

	// Check if all critical components exist
	piperExists := false
	espeakExists := false
	piperDirExists := false

	if _, err := os.Stat(piperExe); err == nil {
		piperExists = true
	}
	if info, err := os.Stat(espeakData); err == nil && info.IsDir() {
		espeakExists = true
	}
	if info, err := os.Stat(piperDir); err == nil && info.IsDir() {
		piperDirExists = true
	}

	// If all dependencies exist, no need to download
	if piperExists && espeakExists && piperDirExists {
		if !obfuscation.Obfuscated() {
			log.Printf("Using cached TTS requirements at: %s", tempRequirementsDir)
		}
		return nil
	}

	if !obfuscation.Obfuscated() {
		log.Printf("Downloading TTS requirements to: %s", tempRequirementsDir)
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempRequirementsDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	var scriptContent []byte
	var scriptPath string
	var cmd *exec.Cmd
	var err error

	if runtime.GOOS == "windows" {
		// Use PowerShell script on Windows
		scriptContent, err = embeddedGetRequirements.ReadFile("get-requirements.ps1")
		if err != nil {
			return fmt.Errorf("failed to read embedded PowerShell script: %w", err)
		}

		// Create a modified version of the script that uses tempRequirementsDir
		modifiedScript := strings.Replace(string(scriptContent),
			`$RequirementsDir = Join-Path $PSScriptRoot "tts\requirements"`,
			`$RequirementsDir = "`+tempRequirementsDir+`"`,
			1)

		// Write the modified script to a temporary file
		scriptPath = filepath.Join(os.TempDir(), "pejelagarto-get-requirements.ps1")
		if err := os.WriteFile(scriptPath, []byte(modifiedScript), 0755); err != nil {
			return fmt.Errorf("failed to write temporary PowerShell script: %w", err)
		}
		defer os.Remove(scriptPath) // Clean up script after execution

		// Execute the PowerShell script
		if !obfuscation.Obfuscated() {
			if singleLanguage != "" {
				log.Printf("Running PowerShell script to download dependencies for language: %s...\n", singleLanguage)
			} else {
				log.Println("Running PowerShell script to download all dependencies...")
			}
		}
		if singleLanguage != "" {
			cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "-Language", singleLanguage)
		} else {
			cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		}
	} else {
		// Use shell script on Linux/macOS
		scriptContent, err = embeddedGetRequirements.ReadFile("get-requirements.sh")
		if err != nil {
			return fmt.Errorf("failed to read embedded shell script: %w", err)
		}

		// Create a modified version of the script that uses tempRequirementsDir
		modifiedScript := strings.Replace(string(scriptContent),
			`REQUIREMENTS_DIR="${SCRIPT_DIR}/tts/requirements"`,
			`REQUIREMENTS_DIR="`+tempRequirementsDir+`"`,
			1)

		// Write the modified script to a temporary file
		scriptPath = filepath.Join("/tmp", "pejelagarto-get-requirements.sh")
		if err := os.WriteFile(scriptPath, []byte(modifiedScript), 0755); err != nil {
			return fmt.Errorf("failed to write temporary shell script: %w", err)
		}
		defer os.Remove(scriptPath) // Clean up script after execution

		// Execute the shell script
		if !obfuscation.Obfuscated() {
			if singleLanguage != "" {
				log.Printf("Running shell script to download dependencies for language: %s...\n", singleLanguage)
			} else {
				log.Println("Running shell script to download all dependencies...")
			}
		}
		if singleLanguage != "" {
			cmd = exec.Command("/bin/bash", scriptPath, singleLanguage)
		} else {
			cmd = exec.Command("/bin/bash", scriptPath)
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute dependency download script: %w", err)
	}

	// Verify that the dependencies were downloaded
	if _, err := os.Stat(piperExe); err != nil {
		return fmt.Errorf("piper binary not found after download: %w", err)
	}
	if _, err := os.Stat(espeakData); err != nil {
		return fmt.Errorf("espeak-ng-data not found after download: %w", err)
	}
	if _, err := os.Stat(piperDir); err != nil {
		return fmt.Errorf("piper directory not found after download: %w", err)
	}

	if !obfuscation.Obfuscated() {
		log.Printf("Successfully downloaded TTS requirements")
	}
	return nil
}

func getPiperBinaryPath() string {
	binaryPath := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	return binaryPath
}

func getModelPath(language string) string {
	modelsDir := filepath.Join(tempRequirementsDir, "piper", "languages", language)
	files, _ := filepath.Glob(filepath.Join(modelsDir, "*.onnx"))
	if len(files) > 0 {
		return files[0]
	}
	return filepath.Join(modelsDir, "model.onnx")
}

func getBaseVowelForTTS(r rune) rune {
	vowelMap := map[rune]rune{
		'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a', 'ă': 'a',
		'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ė': 'e', 'ę': 'e', 'ě': 'e',
		'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ī': 'i', 'į': 'i', 'ı': 'i',
		'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ō': 'o', 'ő': 'o',
		'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ū': 'u', 'ű': 'u', 'ų': 'u',
	}
	if base, ok := vowelMap[r]; ok {
		return base
	}
	return 0
}

func preprocessTextForTTS(input string, pronunciationLanguage string) string {
	// Convert numbers from Pejelagarto format
	input = applyNumbersLogicFromPejelagarto(input)

	var vowels, consonants, allowed string

	switch pronunciationLanguage {
	case "portuguese":
		vowels = "aeiouáéíóúâêôãõàü"
		consonants = "bcdfghjklmnpqrstvwxyzç"
		allowed = vowels + consonants + " "
	case "french":
		vowels = "aeiouyàâäæçéèêëîïôœùûüÿ"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + " "
	case "russian":
		vowels = "аеёиоуыэюя"
		consonants = "бвгджзйклмнпрстфхцчшщъьї"
		allowed = vowels + consonants + " "
	case "german":
		vowels = "aeiouyäöü"
		consonants = "bcdfghjklmnpqrstvwxzß"
		allowed = vowels + consonants + " "
	case "hindi":
		// Hindi uses Devanagari script
		return input
	case "romanian":
		vowels = "aeiouăâî"
		consonants = "bcdfghjklmnpqrstvwxyzțș"
		allowed = vowels + consonants + " "
	case "swahili":
		vowels = "aeiou"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + " "
	case "czech":
		vowels = "aeiouyáéíóúůýě"
		consonants = "bcčdďfghjklmnňpqrřsštťvwxzž"
		allowed = vowels + consonants + " "
	case "icelandic":
		vowels = "aeiouyáéíóúýæöþð"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + " "
	case "kazakh":
		// Kazakh uses Cyrillic
		vowels = "аәеёиоөұүыэюя"
		consonants = "бвгғджзйкқлмнңопрстуфхһцчшщъыьіэюя"
		allowed = vowels + consonants + " "
	case "norwegian":
		vowels = "aeiouyæøåäöü"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + " "
	case "swedish":
		vowels = "aeiouyåäö"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + " "
	case "turkish":
		vowels = "aeıioöuüâî"
		consonants = "bcçdfgğhjklmnprsştvyzw"
		allowed = vowels + consonants + " "
	case "vietnamese":
		vowels = "aăâeêioôơuưy"
		consonants = "bcđdfghjklmnpqrstvxz"
		allowed = vowels + consonants + " "
	case "hungarian":
		vowels = "aáeéiíoóöőuúüű"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + " "
	case "chinese":
		// Chinese uses Han characters - minimal preprocessing
		return input
	default:
		vowels = "aeiou"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + " "
	}

	vowels = strings.ToLower(vowels) + strings.ToUpper(vowels)
	consonants = strings.ToLower(consonants) + strings.ToUpper(consonants)
	allowed = strings.ToLower(allowed) + strings.ToUpper(allowed)

	var result []rune
	consecutiveConsonants := 0

	for _, r := range input {
		lowerR := unicode.ToLower(r)

		if baseVowel := getBaseVowelForTTS(lowerR); baseVowel != 0 {
			if !strings.ContainsRune(vowels, baseVowel) {
				if unicode.IsUpper(r) {
					r = unicode.ToUpper(baseVowel)
				} else {
					r = baseVowel
				}
			}
		}

		if !strings.ContainsRune(allowed, r) && r != ' ' {
			continue
		}

		if strings.ContainsRune(consonants, r) {
			consecutiveConsonants++
			if consecutiveConsonants > 2 {
				continue
			}
		} else {
			consecutiveConsonants = 0
		}

		result = append(result, r)
	}

	output := string(result)
	if strings.TrimSpace(output) == "" {
		return "..."
	}
	return output
}

func textToSpeech(input string, pronunciationLanguage string) (outputPath string, err error) {
	input = preprocessTextForTTS(input, pronunciationLanguage)

	modelPath := getModelPath(pronunciationLanguage)
	binaryPath := getPiperBinaryPath()

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("piper binary not found at %s: %w", binaryPath, err)
	}

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("voice model not found at %s: %w", modelPath, err)
	}

	tempFile, err := os.CreateTemp("", "piper-tts-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	tempFile.Close()

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for output: %w", err)
	}
	absModelPath, err := filepath.Abs(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for model: %w", err)
	}

	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for binary: %w", err)
	}

	absRequirementsDir := tempRequirementsDir

	cmd := exec.Command(
		absBinaryPath,
		"-m", absModelPath,
		"--output_file", absOutputPath,
	)

	cmd.Dir = absRequirementsDir
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("piper command failed: %w\nOutput: %s", err, string(output))
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("output file not created: %w", err)
	}
	if fileInfo.Size() == 0 {
		os.Remove(outputPath)
		return "", fmt.Errorf("output file is empty (piper output: %s)", string(output))
	}

	return outputPath, nil
}

func slowDownAudio(inputPath string) (outputPath string, err error) {
	tempFile, err := os.CreateTemp("", "piper-tts-slow-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	tempFile.Close()

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-filter:a", "atempo=0.5",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("ffmpeg command failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

func handleTextToSpeech(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pronunciationLanguage
	}

	slow := r.URL.Query().Get("slow") == "true"

	validLanguages := map[string]bool{
		"russian": true, "portuguese": true, "french": true, "german": true,
		"hindi": true, "romanian": true, "swahili": true, "czech": true,
		"icelandic": true, "kazakh": true, "norwegian": true, "swedish": true,
		"turkish": true, "vietnamese": true, "hungarian": true, "chinese": true,
	}
	if !validLanguages[lang] {
		http.Error(w, fmt.Sprintf("Invalid language '%s'", lang), http.StatusBadRequest)
		return
	}

	input := string(body)

	cacheKey := fmt.Sprintf("%s:%s:%v", input, lang, slow)

	audioCache.RLock()
	cachedAudio, exists := audioCache.cache[cacheKey]
	audioCache.RUnlock()

	if exists {
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Content-Disposition", "inline")
		w.Write(cachedAudio)
		return
	}

	wavPath, err := textToSpeech(input, lang)
	if err != nil {
		http.Error(w, fmt.Sprintf("TTS error: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(wavPath)

	if slow {
		slowWavPath, err := slowDownAudio(wavPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Audio slowdown error: %v", err), http.StatusInternalServerError)
			return
		}
		defer os.Remove(slowWavPath)
		wavPath = slowWavPath
	}

	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		http.Error(w, "Error reading audio file", http.StatusInternalServerError)
		return
	}

	audioCache.Lock()
	audioCache.cache[cacheKey] = wavData
	audioCache.Unlock()

	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "inline")
	w.Write(wavData)
}

func handleCheckSlowAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pronunciationLanguage
	}

	input := string(body)
	cacheKey := fmt.Sprintf("%s:%s:true", input, lang)

	audioCache.RLock()
	_, exists := audioCache.cache[cacheKey]
	audioCache.RUnlock()

	if exists {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ready":true}`)
		return
	}

	go func() {
		audioCache.RLock()
		_, exists := audioCache.cache[cacheKey]
		audioCache.RUnlock()
		if exists {
			return
		}

		wavPath, err := textToSpeech(input, lang)
		if err != nil {
			return
		}
		defer os.Remove(wavPath)

		slowWavPath, err := slowDownAudio(wavPath)
		if err != nil {
			return
		}
		defer os.Remove(slowWavPath)

		slowData, err := os.ReadFile(slowWavPath)
		if err != nil {
			return
		}

		audioCache.Lock()
		audioCache.cache[cacheKey] = slowData
		audioCache.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ready":false}`)
}

// applyNumbersLogicFromPejelagarto is needed by preprocessTextForTTS
func applyNumbersLogicFromPejelagarto(input string) string {
	// Pattern to match base-8 sequences for positive numbers (prefixed with optional +, not -)
	base8Pattern := regexp.MustCompile(`(?:^|[^-])\+?([0-7]+)`)

	// Pattern to match base-7 sequences for negative numbers (prefixed with -)
	base7Pattern := regexp.MustCompile(`-([0-6]+)`)

	result := input

	// Process positive numbers (base-8)
	result = base8Pattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the number part
		num := strings.TrimLeft(match, "+")
		num = strings.TrimLeft(num, " ")

		// Skip if contains digits 8 or 9 (not base-8)
		if strings.ContainsAny(num, "89") {
			return match
		}

		// Convert from base-8 to base-10
		val := int64(0)
		for _, digit := range num {
			val = val*8 + int64(digit-'0')
		}

		return fmt.Sprintf("%d", val)
	})

	// Process negative numbers (base-7)
	result = base7Pattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the number part (without the minus sign)
		num := strings.TrimPrefix(match, "-")

		// Skip if contains digits 7, 8, or 9 (not base-7)
		if strings.ContainsAny(num, "789") {
			return match
		}

		// Convert from base-7 to base-10
		val := int64(0)
		for _, digit := range num {
			val = val*7 + int64(digit-'0')
		}

		return fmt.Sprintf("-%d", val)
	})

	return result
}
