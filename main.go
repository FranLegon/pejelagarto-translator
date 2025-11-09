package main

import (
	"embed"
)

//go:embed scripts/requirements/get-requirements.ps1 scripts/requirements/get-requirements.sh
var embeddedGetRequirements embed.FS

const htmlUI = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pejelagarto Translator</title>
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
            <a href="/download/android" download="pejelagarto-translator.apk" class="download-btn">
                📱 Android
            </a>
        </div>
    </div>
    
    <script>
        let isInverted = false;
        let liveTranslateEnabled = true;
        
        // Initialize theme on page load
        (function initTheme() {
            // Check localStorage for saved preference, default to dark mode
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
        
        // Toggle theme function
        function toggleTheme() {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            updateThemeIcon(newTheme);
        }
        
        // Update theme icon
        function updateThemeIcon(theme) {
            const icon = document.getElementById('theme-icon');
            icon.textContent = theme === 'dark' ? '🌙' : '☀️';
        }
        
        // Handle translate button click
        function handleTranslateClick() {
            handleLiveTranslation();
        }
        
        // Invert button functionality
        function invertTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const translateBtn = document.getElementById('translate-btn');
            
            // Swap text content
            const temp = inputText.value;
            inputText.value = outputText.value;
            outputText.value = temp;
            
            // Toggle inverted state
            isInverted = !isInverted;
            
            // Update translate button text
            if (isInverted) {
                translateBtn.textContent = 'Translate from Pejelagarto';
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
            }
            
            // Force reset to single button state with proper labels
            resetToSingleButton();
        }
        
        // Live translation functionality
        function toggleLiveTranslation() {
            const checkbox = document.getElementById('live-translate');
            const translateBtn = document.getElementById('translate-btn');
            const inputText = document.getElementById('input-text');
            
            liveTranslateEnabled = checkbox.checked;
            
            if (liveTranslateEnabled) {
                // Hide the translate button
                translateBtn.classList.add('hidden');
                
                // Add live translation event listener
                inputText.addEventListener('input', handleLiveTranslation);
                
                // Trigger initial translation
                handleLiveTranslation();
            } else {
                // Show the translate button
                translateBtn.classList.remove('hidden');
                
                // Remove live translation event listener
                inputText.removeEventListener('input', handleLiveTranslation);
            }
        }
        
        function handleLiveTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const endpoint = isInverted ? '/from' : '/to';
            
            // Send request to backend
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain'
                },
                body: inputText.value
            })
            .then(response => response.text())
            .then(data => {
                outputText.value = data;
            })
            .catch(error => {
                console.error('Translation error:', error);
            });
        }
        
        // Track current output text, language, and slow audio availability
        let currentOutputText = '';
        let currentLanguage = '';
        let currentInvertedState = false;
        let slowAudioReady = {}; // Tracks which text+language combinations have slow audio ready
        
        // Watch for output text changes or language changes to reset buttons
        function watchOutputChanges() {
            const outputText = document.getElementById('output-text');
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            // Check if state has changed (text, language, or invert state)
            if (outputText.value !== currentOutputText || selectedLanguage !== currentLanguage || isInverted !== currentInvertedState) {
                currentOutputText = outputText.value;
                currentLanguage = selectedLanguage;
                currentInvertedState = isInverted;
                resetToSingleButton();
                
                // Check if slow audio is already ready for this text+language combination
                const cacheKey = currentOutputText + ':' + selectedLanguage;
                if (slowAudioReady[cacheKey]) {
                    // Already have slow audio, split button immediately
                    const source = isInverted ? 'input' : 'output';
                    const container = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
                    splitButton(source, container);
                }
            }
        }
        
        // Reset to single button state
        function resetToSingleButton() {
            const outputLabel = document.getElementById('output-label');
            const inputLabel = document.getElementById('input-label');
            
            // Get current language selection before recreating dropdown
            const oldDropdown = document.getElementById('tts-language');
            const selectedLang = oldDropdown ? oldDropdown.value : 'russian';
            
            const dropdownHTML = document.getElementById('tts-language') ? 
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="kazakh">North-North-East</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">South-East-East</option><option value="hindi">South-East</option><option value="icelandic">South-South-East</option><option value="romanian">South</option><option value="vietnamese">South-South-West</option><option value="swahili">South-West</option><option value="swedish">South-West-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option></select>' : '';
            
            // Always reset both labels to ensure clean state
            if (isInverted) {
                inputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-input" onclick="playAudio(&quot;input&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                outputLabel.textContent = 'Human:';
            } else {
                outputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-output" onclick="playAudio(&quot;output&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                inputLabel.textContent = 'Human:';
            }
            
            // Restore language selection
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
        
        // Start watching for changes
        setInterval(watchOutputChanges, 500);
        
        // Play audio function - only called when play button is clicked
        function playAudio(source, slow) {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const playInputBtn = document.getElementById('play-input');
            const playOutputBtn = document.getElementById('play-output');
            const playInputSlowBtn = document.getElementById('play-input-slow');
            const playOutputSlowBtn = document.getElementById('play-output-slow');
            
            // Determine which text to convert to speech
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
            
            // Check if there's text to speak
            if (!textToSpeak || textToSpeak.trim() === '') {
                alert('No text to convert to speech!');
                return;
            }
            
            // Disable button and show loading state
            button.disabled = true;
            const originalText = button.textContent;
            button.textContent = '⏳';
            
            // Get selected language from dropdown if available
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            // Build URL with language parameter and slow parameter
            let url = selectedLanguage ? '/tts?lang=' + selectedLanguage : '/tts';
            if (slow) {
                url += (selectedLanguage ? '&' : '?') + 'slow=true';
            }
            
            // Send request to TTS endpoint
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
                // Create an audio element and play it
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
                
                // If this is normal speed, start checking for slow version
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
        
        // Check periodically if slow audio is ready
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
                        // Mark this text+language combination as ready
                        slowAudioReady[cacheKey] = true;
                        splitButton(source, container);
                    }
                })
                .catch(error => {
                    console.error('Error checking slow audio:', error);
                    clearInterval(checkInterval);
                });
            }, 1000); // Check every second
            
            // Stop checking after 30 seconds
            setTimeout(() => clearInterval(checkInterval), 30000);
        }
        
        // Split button into fast and slow versions
        function splitButton(source, container) {
            // Validate that we're modifying the correct container based on current state
            const expectedSource = isInverted ? 'input' : 'output';
            if (source !== expectedSource) {
                // State has changed since this was called, don't modify
                return;
            }
            
            // Double-check the container is the correct label
            const expectedContainer = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
            if (container !== expectedContainer) {
                // Container mismatch, state has changed
                return;
            }
            
            // Get current language selection before recreating dropdown
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
            
            // Restore language selection
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
    </script>
    
    <div class="version-display"><a href="https://github.com/FranLegon/pejelagarto-translator" target="_blank">{{VERSION}}</a></div>
</body>
</html>`

// HTTP handler for the main UI
