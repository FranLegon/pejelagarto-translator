package com.pejelagarto.translator;

import android.os.Bundle;
import android.util.Log;
import android.webkit.ConsoleMessage;
import android.webkit.JavascriptInterface;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import androidx.appcompat.app.AppCompatActivity;

import translator.Translator;

public class MainActivity extends AppCompatActivity {
    
    private WebView webView;
    private static final String TAG = "PejeLagartoApp";
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        
        webView = findViewById(R.id.webview);
        WebSettings webSettings = webView.getSettings();
        webSettings.setJavaScriptEnabled(true);
        webSettings.setDomStorageEnabled(true);
        webSettings.setCacheMode(WebSettings.LOAD_NO_CACHE);
        
        // Add console logging
        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onConsoleMessage(ConsoleMessage msg) {
                Log.d(TAG, "WebView Console: " + msg.message() + " at " + msg.lineNumber());
                return true;
            }
        });
        
        // Add JavaScript interface to call Go functions
        webView.addJavascriptInterface(new TranslatorBridge(), "AndroidTranslator");
        
        // Clear cache and load the HTML content
        webView.clearCache(true);
        webView.loadDataWithBaseURL("file:///android_asset/", getHtmlContent(), "text/html", "UTF-8", null);
    }
    
    private class TranslatorBridge {
        @JavascriptInterface
        public String translateToPejelagarto(String text) {
            try {
                // Use Translator_ class instance methods from gomobile bind
                translator.Translator_ t = translator.Translator.new_();
                return t.translateToPejelagarto(text);
            } catch (Exception e) {
                return "Error: " + e.getMessage();
            }
        }
        
        @JavascriptInterface
        public String translateFromPejelagarto(String text) {
            try {
                // Use Translator_ class instance methods from gomobile bind
                translator.Translator_ t = translator.Translator.new_();
                return t.translateFromPejelagarto(text);
            } catch (Exception e) {
                return "Error: " + e.getMessage();
            }
        }
    }
    
    private String getHtmlContent() {
        return "<!DOCTYPE html>\n" +
            "<html>\n" +
            "<head>\n" +
            "    <meta charset=\"UTF-8\">\n" +
            "    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no\">\n" +
            "    <title>Pejelagarto Translator</title>\n" +
            "    <style>\n" +
            "        * {\n" +
            "            margin: 0;\n" +
            "            padding: 0;\n" +
            "            box-sizing: border-box;\n" +
            "        }\n" +
            "        body {\n" +
            "            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;\n" +
            "            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);\n" +
            "            min-height: 100vh;\n" +
            "            padding: 20px;\n" +
            "            display: flex;\n" +
            "            flex-direction: column;\n" +
            "            align-items: center;\n" +
            "            justify-content: center;\n" +
            "        }\n" +
            "        .container {\n" +
            "            width: 100%;\n" +
            "            max-width: 600px;\n" +
            "            background: white;\n" +
            "            border-radius: 20px;\n" +
            "            padding: 30px;\n" +
            "            box-shadow: 0 20px 60px rgba(0,0,0,0.3);\n" +
            "        }\n" +
            "        h1 {\n" +
            "            color: #667eea;\n" +
            "            text-align: center;\n" +
            "            margin-bottom: 10px;\n" +
            "            font-size: 28px;\n" +
            "        }\n" +
            "        .subtitle {\n" +
            "            text-align: center;\n" +
            "            color: #999;\n" +
            "            font-size: 14px;\n" +
            "            margin-bottom: 30px;\n" +
            "        }\n" +
            "        .input-group {\n" +
            "            margin-bottom: 20px;\n" +
            "        }\n" +
            "        label {\n" +
            "            display: block;\n" +
            "            margin-bottom: 8px;\n" +
            "            color: #666;\n" +
            "            font-weight: 600;\n" +
            "        }\n" +
            "        textarea {\n" +
            "            width: 100%;\n" +
            "            padding: 15px;\n" +
            "            border: 2px solid #e0e0e0;\n" +
            "            border-radius: 10px;\n" +
            "            font-size: 16px;\n" +
            "            resize: vertical;\n" +
            "            min-height: 120px;\n" +
            "            font-family: inherit;\n" +
            "        }\n" +
            "        textarea:focus {\n" +
            "            outline: none;\n" +
            "            border-color: #667eea;\n" +
            "        }\n" +
            "        .button-group {\n" +
            "            display: flex;\n" +
            "            gap: 10px;\n" +
            "            margin-bottom: 20px;\n" +
            "        }\n" +
            "        button {\n" +
            "            flex: 1;\n" +
            "            padding: 15px;\n" +
            "            border: none;\n" +
            "            border-radius: 10px;\n" +
            "            font-size: 16px;\n" +
            "            font-weight: 600;\n" +
            "            cursor: pointer;\n" +
            "            transition: all 0.3s;\n" +
            "        }\n" +
            "        .btn-primary {\n" +
            "            background: #667eea;\n" +
            "            color: white;\n" +
            "        }\n" +
            "        .btn-primary:active {\n" +
            "            background: #5568d3;\n" +
            "            transform: scale(0.98);\n" +
            "        }\n" +
            "        .btn-secondary {\n" +
            "            background: #f0f0f0;\n" +
            "            color: #333;\n" +
            "        }\n" +
            "        .btn-secondary:active {\n" +
            "            background: #e0e0e0;\n" +
            "            transform: scale(0.98);\n" +
            "        }\n" +
            "        .output {\n" +
            "            background: #f9f9f9;\n" +
            "            padding: 15px;\n" +
            "            border-radius: 10px;\n" +
            "            min-height: 120px;\n" +
            "            white-space: pre-wrap;\n" +
            "            word-wrap: break-word;\n" +
            "            border: 2px solid #e0e0e0;\n" +
            "            font-size: 16px;\n" +
            "        }\n" +
            "        .swap-btn {\n" +
            "            width: 100%;\n" +
            "            margin: 10px 0;\n" +
            "            background: #764ba2;\n" +
            "        }\n" +
            "    </style>\n" +
            "</head>\n" +
            "<body>\n" +
            "    <div class=\"container\">\n" +
            "        <h1>🦎 Pejelagarto Translator</h1>\n" +
            "        <div class=\"subtitle\">Native Android App</div>\n" +
            "        \n" +
            "        <div class=\"input-group\">\n" +
            "            <label id=\"inputLabel\">English:</label>\n" +
            "            <textarea id=\"inputText\" placeholder=\"Type your text here...\"></textarea>\n" +
            "        </div>\n" +
            "        \n" +
            "        <div class=\"button-group\">\n" +
            "            <button class=\"btn-primary\" onclick=\"doTranslate()\">Translate</button>\n" +
            "            <button class=\"btn-secondary\" onclick=\"clearText()\">Clear</button>\n" +
            "        </div>\n" +
            "        \n" +
            "        <button class=\"btn-primary swap-btn\" onclick=\"swapDirection()\">⇅ Swap Direction</button>\n" +
            "        \n" +
            "        <div class=\"input-group\">\n" +
            "            <label id=\"outputLabel\">Pejelagarto:</label>\n" +
            "            <div class=\"output\" id=\"output\">Translation will appear here...</div>\n" +
            "        </div>\n" +
            "    </div>\n" +
            "\n" +
            "    <script>\n" +
            "        let direction = 'toPejelagarto';\n" +
            "\n" +
            "        function doTranslate() {\n" +
            "            const input = document.getElementById('inputText').value;\n" +
            "            if (!input.trim()) {\n" +
            "                document.getElementById('output').textContent = 'Please enter some text to translate.';\n" +
            "                return;\n" +
            "            }\n" +
            "\n" +
            "            try {\n" +
            "                let result;\n" +
            "                if (direction === 'toPejelagarto') {\n" +
            "                    result = AndroidTranslator.translateToPejelagarto(input);\n" +
            "                } else {\n" +
            "                    result = AndroidTranslator.translateFromPejelagarto(input);\n" +
            "                }\n" +
            "                document.getElementById('output').textContent = result;\n" +
            "            } catch (error) {\n" +
            "                document.getElementById('output').textContent = 'Error: ' + error.message;\n" +
            "            }\n" +
            "        }\n" +
            "\n" +
            "        function clearText() {\n" +
            "            document.getElementById('inputText').value = '';\n" +
            "            document.getElementById('output').textContent = 'Translation will appear here...';\n" +
            "        }\n" +
            "\n" +
            "        function swapDirection() {\n" +
            "            if (direction === 'toPejelagarto') {\n" +
            "                direction = 'fromPejelagarto';\n" +
            "                document.getElementById('inputLabel').textContent = 'Pejelagarto:';\n" +
            "                document.getElementById('outputLabel').textContent = 'English:';\n" +
            "            } else {\n" +
            "                direction = 'toPejelagarto';\n" +
            "                document.getElementById('inputLabel').textContent = 'English:';\n" +
            "                document.getElementById('outputLabel').textContent = 'Pejelagarto:';\n" +
            "            }\n" +
            "            clearText();\n" +
            "        }\n" +
            "    </script>\n" +
            "</body>\n" +
            "</html>";
    }
}
