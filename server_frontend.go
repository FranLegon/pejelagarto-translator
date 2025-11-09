//go:build frontendserver
// +build frontendserver

// Simple HTTP server for frontend mode
// This serves the WASM-enabled UI and TTS endpoints only
// Translation happens client-side in the browser via WASM

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"

	"pejelagarto-translator/obfuscation"
)

// Downloadable feature - constants come from downloadable.go or not_downloadable.go based on build tags
// var embeddedBinaries embed.FS and const isDownloadable are defined in those files

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
        
        .pronunciation-container {
            margin-top: 20px;
            display: flex;
            flex-direction: column;
            align-items: center;
            max-width: 800px;
            margin-left: auto;
            margin-right: auto;
        }
        
        .pronunciation-container label {
            text-align: center;
            margin-bottom: 10px;
        }
        
        .pronunciation-container textarea {
            width: 100%;
            height: 120px;
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
        
        .version-display {
            position: fixed;
            bottom: 10px;
            right: 10px;
            font-size: 12px;
            color: var(--text-secondary);
            opacity: 0.7;
            font-family: 'Courier New', monospace;
            z-index: 1000;
        }
        
        .version-display a {
            color: var(--text-secondary);
            text-decoration: none;
            transition: opacity 0.2s ease;
        }
        
        .version-display a:hover {
            opacity: 1;
            text-decoration: underline;
        }
        
        @media (max-width: 768px) {
            .version-display {
                font-size: 10px;
                bottom: 5px;
                right: 5px;
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
        
        <div class="pronunciation-container">
            <label>Pronunciation:</label>
            <textarea id="pronunciation-text" readonly placeholder="Pronunciation will appear here..."></textarea>
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
            
            // Update pronunciation after inversion
            if (outputText.value) {
                updatePronunciation(outputText.value);
            }
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
            const pronunciationText = document.getElementById('pronunciation-text');
            
            try {
                if (!isInverted) {
                    // Human to Pejelagarto
                    outputText.value = GoTranslateToPejelagarto(inputText.value);
                } else {
                    // Pejelagarto to Human
                    outputText.value = GoTranslateFromPejelagarto(inputText.value);
                }
                
                // Update pronunciation
                updatePronunciation(outputText.value);
            } catch (error) {
                console.error('Translation error:', error);
            }
        }
        
        function updatePronunciation(text) {
            const pronunciationText = document.getElementById('pronunciation-text');
            const languageDropdown = document.getElementById('tts-language');
            const lang = languageDropdown ? languageDropdown.value : '';
            
            if (!text) {
                pronunciationText.value = '';
                return;
            }
            
            fetch('/pronunciation?lang=' + encodeURIComponent(lang), {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain; charset=utf-8'
                },
                body: text
            })
            .then(response => response.text())
            .then(pronunciation => {
                pronunciationText.value = pronunciation;
            })
            .catch(error => {
                console.error('Pronunciation error:', error);
                pronunciationText.value = '';
            });
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
                
                // Update pronunciation when language changes
                if (outputText.value) {
                    updatePronunciation(outputText.value);
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
    
    <div class="version-display"><a href="https://github.com/FranLegon/pejelagarto-translator" target="_blank">{{VERSION}}</a></div>
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

	// Replace version placeholder
	html = strings.Replace(html, "{{VERSION}}", Version, 1)

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

// findAvailablePort checks if a port is available and returns it, or tries fallbacks
func findAvailablePort() int {
	// Primary port and fallback list
	ports := []int{8080, 8081, 8082, 8083, 8084, 8085, 8086, 8087, 8088, 8089, 8090}

	for _, port := range ports {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port
		}
	}

	// If all ports are taken, return 0 to let the system assign one
	return 0
}

func main() {
	// Parse command-line flags
	var ngrokToken *string
	var ngrokDomain *string

	if useNgrokDefault {
		// Use hardcoded values for ngrok_default builds
		token := defaultNgrokToken
		domain := defaultNgrokDomain
		ngrokToken = &token
		ngrokDomain = &domain
		if !obfuscation.Obfuscated() {
			log.Println("Using hardcoded ngrok configuration (ngrok_default build)")
		}
	} else {
		// Use command-line flags for regular builds
		ngrokToken = flag.String("ngrok_token", "", "Optional ngrok auth token to expose server publicly")
		ngrokDomain = flag.String("ngrok_domain", "", "Optional ngrok persistent domain (e.g., your-domain.ngrok-free.app)")
	}

	pronunciationLangFlag := flag.String("pronunciation_language", "russian", "TTS pronunciation language")
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, "Show language dropdown in UI for TTS")
	flag.Parse()

	if *ngrokDomain != "" && !strings.HasPrefix(*ngrokDomain, "http://") && !strings.HasPrefix(*ngrokDomain, "https://") {
		*ngrokDomain = "https://" + *ngrokDomain
	}

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
	http.HandleFunc("/pronunciation", handlePronunciation)

	// Download endpoints
	http.HandleFunc("/api/is-downloadable", handleIsDownloadable)
	http.HandleFunc("/download/windows", handleDownloadWindows)
	http.HandleFunc("/download/linux", handleDownloadLinux)

	if *ngrokToken != "" {
		// Use ngrok to expose server publicly
		if !obfuscation.Obfuscated() {
			log.Println("Initializing ngrok tunnel...")
			log.Printf("Using auth token: %s...\n", (*ngrokToken)[:10])
			log.Println("Connecting to ngrok service...")
		}

		// Configure endpoint with optional domain
		var listener ngrok.Tunnel
		var err error

		if *ngrokDomain != "" {
			// Strip scheme from domain for WithDomain (it expects just hostname)
			domain := *ngrokDomain
			domain = strings.TrimPrefix(domain, "https://")
			domain = strings.TrimPrefix(domain, "http://")

			if !obfuscation.Obfuscated() {
				log.Printf("Using persistent domain: %s\n", domain)
				log.Println("Establishing tunnel (this may take a few seconds)...")
			}

			// Use a channel to receive the result with timeout
			type result struct {
				listener ngrok.Tunnel
				err      error
			}
			resultChan := make(chan result)
			go func() {
				l, e := ngrok.Listen(context.Background(),
					config.HTTPEndpoint(
						config.WithDomain(domain),
					),
					ngrok.WithAuthtoken(*ngrokToken),
				)
				resultChan <- result{listener: l, err: e}
			}()

			// Wait for completion or timeout (increased to 45 seconds for slower connections)
			select {
			case res := <-resultChan:
				listener = res.listener
				err = res.err
			case <-time.After(45 * time.Second):
				log.Fatalf("Failed to start ngrok listener: connection timeout after 45 seconds.\n\nPossible causes:\n  - Slow internet connection\n  - ngrok service unavailable\n  - Domain '%s' may be in use\n\nTry:\n  - Check internet connectivity\n  - Run without -ngrok_domain to use random URL\n  - Wait a few minutes and retry", domain)
			}
		} else {
			if !obfuscation.Obfuscated() {
				log.Println("Using random ngrok domain")
				log.Println("Establishing tunnel (this may take a few seconds)...")
			}

			// Use a channel to receive the result with timeout
			type result struct {
				listener ngrok.Tunnel
				err      error
			}
			resultChan := make(chan result)
			go func() {
				l, e := ngrok.Listen(context.Background(),
					config.HTTPEndpoint(),
					ngrok.WithAuthtoken(*ngrokToken),
				)
				resultChan <- result{listener: l, err: e}
			}()

			// Wait for completion or timeout
			select {
			case res := <-resultChan:
				listener = res.listener
				err = res.err
			case <-time.After(10 * time.Second):
				log.Fatalf("Failed to start ngrok listener: connection timeout after 10 seconds")
			}
		}

		if err != nil {
			// Check for specific error types and provide helpful messages
			errStr := err.Error()
			if strings.Contains(errStr, "already online") || strings.Contains(errStr, "ERR_NGROK_334") {
				log.Fatalf("Failed to start ngrok listener: The domain '%s' is already in use.\nThis could mean:\n  1. Another instance is using this domain\n  2. A previous tunnel wasn't properly closed\n\nPlease either:\n  - Stop the other instance using this domain\n  - Wait a few minutes for the old tunnel to expire\n  - Use a different domain\n\nError: %v", *ngrokDomain, err)
			} else if strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "invalid authtoken") {
				log.Fatalf("Failed to start ngrok listener: Invalid authentication token.\nPlease check your ngrok auth token.\n\nError: %v", err)
			} else {
				log.Fatalf("Failed to start ngrok listener: %v\n\nTroubleshooting:\n  - Check your internet connection\n  - Verify ngrok service is available\n  - Try running without a fixed domain", err)
			}
		}

		url := listener.URL()
		if !obfuscation.Obfuscated() {
			log.Printf("ngrok tunnel established successfully! ✓\n")
			log.Printf("Public URL: %s\n", url)
		}

		// Open browser with ngrok URL (only if configured to do so)
		if obfuscation.ShouldOpenBrowser() {
			go func() {
				time.Sleep(1 * time.Second)
				if err := openBrowser(url); err != nil {
					if !obfuscation.Obfuscated() {
						log.Printf("Could not open browser automatically: %v\n", err)
						log.Printf("Please open your browser and navigate to %s\n", url)
					}
				}
			}()
		}

		log.Println("Server is running. Press Ctrl+C to stop.")
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		// Use local server (default behavior)
		port := findAvailablePort()
		if port == 0 {
			log.Fatal("No available ports found in range 8080-8090")
		}
		addr := fmt.Sprintf(":%d", port)
		url := fmt.Sprintf("http://localhost:%d", port)

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
