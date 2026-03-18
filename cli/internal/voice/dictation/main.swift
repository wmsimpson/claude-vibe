import Foundation
import Speech
import AVFoundation

/// Dictation captures microphone audio and performs real-time speech recognition
/// using macOS native SFSpeechRecognizer. Results are streamed as JSON lines to stdout.
///
/// Output format (one JSON object per line):
///   {"type":"ready","text":""}           - Microphone ready, listening
///   {"type":"partial","text":"hello"}    - Interim transcription
///   {"type":"final","text":"hello world"} - Final transcription (on silence timeout)
///   {"type":"error","text":"..."}        - Error message
class Dictation {
    let recognizer: SFSpeechRecognizer
    let audioEngine = AVAudioEngine()
    var request: SFSpeechAudioBufferRecognitionRequest?
    var task: SFSpeechRecognitionTask?
    var silenceTimer: DispatchSourceTimer?
    let silenceTimeout: TimeInterval
    var lastTranscript = ""
    var hasSpeech = false

    init(locale: String, silenceTimeout: TimeInterval) {
        guard let rec = SFSpeechRecognizer(locale: Locale(identifier: locale)) else {
            Dictation.emit(type: "error", text: "Speech recognizer not available for locale: \(locale)")
            exit(1)
        }
        self.recognizer = rec
        self.silenceTimeout = silenceTimeout
    }

    func start() {
        SFSpeechRecognizer.requestAuthorization { status in
            switch status {
            case .authorized:
                DispatchQueue.main.async { self.beginRecognition() }
            case .denied:
                Dictation.emit(type: "error", text: "Speech recognition denied. Enable in System Settings > Privacy & Security > Speech Recognition.")
                exit(1)
            case .restricted:
                Dictation.emit(type: "error", text: "Speech recognition restricted on this device.")
                exit(1)
            case .notDetermined:
                Dictation.emit(type: "error", text: "Speech recognition permission not determined.")
                exit(1)
            @unknown default:
                Dictation.emit(type: "error", text: "Unknown authorization status.")
                exit(1)
            }
        }
    }

    func beginRecognition() {
        request = SFSpeechAudioBufferRecognitionRequest()
        guard let request = request else {
            Dictation.emit(type: "error", text: "Failed to create recognition request.")
            exit(1)
        }

        request.shouldReportPartialResults = true
        if #available(macOS 13, *) {
            if recognizer.supportsOnDeviceRecognition {
                request.requiresOnDeviceRecognition = true
            }
        }

        let node = audioEngine.inputNode
        let format = node.outputFormat(forBus: 0)

        node.installTap(onBus: 0, bufferSize: 1024, format: format) { [weak self] buffer, _ in
            self?.request?.append(buffer)
        }

        task = recognizer.recognitionTask(with: request) { [weak self] result, error in
            guard let self = self else { return }

            if let result = result {
                let text = result.bestTranscription.formattedString
                let isFinal = result.isFinal

                if text != self.lastTranscript {
                    self.lastTranscript = text
                    self.hasSpeech = true
                    Dictation.emit(type: isFinal ? "final" : "partial", text: text)
                    self.resetSilenceTimer()
                }

                if isFinal {
                    self.cleanup()
                    exit(0)
                }
            }

            if let error = error {
                let nsError = error as NSError
                // Code 209/216 = recognition cancelled/ended, expected on stop
                if nsError.domain == "kAFAssistantErrorDomain" &&
                    (nsError.code == 209 || nsError.code == 216) {
                    return
                }
                Dictation.emit(type: "error", text: error.localizedDescription)
                self.cleanup()
                exit(1)
            }
        }

        audioEngine.prepare()
        do {
            try audioEngine.start()
        } catch {
            Dictation.emit(type: "error", text: "Failed to start audio engine: \(error.localizedDescription)")
            exit(1)
        }

        Dictation.emit(type: "ready", text: "")
    }

    func resetSilenceTimer() {
        silenceTimer?.cancel()
        let timer = DispatchSource.makeTimerSource(queue: .main)
        timer.schedule(deadline: .now() + silenceTimeout)
        timer.setEventHandler { [weak self] in
            self?.finalize()
        }
        timer.resume()
        silenceTimer = timer
    }

    func finalize() {
        if hasSpeech && !lastTranscript.isEmpty {
            Dictation.emit(type: "final", text: lastTranscript)
        }
        cleanup()
        exit(0)
    }

    func cleanup() {
        silenceTimer?.cancel()
        audioEngine.stop()
        audioEngine.inputNode.removeTap(onBus: 0)
        request?.endAudio()
        task?.cancel()
    }

    /// Emit a JSON event to stdout. Thread-safe via synchronous write.
    static func emit(type: String, text: String) {
        let escaped = text
            .replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "\"", with: "\\\"")
            .replacingOccurrences(of: "\n", with: "\\n")
            .replacingOccurrences(of: "\r", with: "\\r")
            .replacingOccurrences(of: "\t", with: "\\t")
        let json = "{\"type\":\"\(type)\",\"text\":\"\(escaped)\"}\n"
        if let data = json.data(using: .utf8) {
            FileHandle.standardOutput.write(data)
        }
    }
}

// --- Main ---

var silenceTimeout = 3.0
var locale = "en-US"

let args = CommandLine.arguments
var i = 1
while i < args.count {
    switch args[i] {
    case "--timeout":
        if i + 1 < args.count, let t = Double(args[i + 1]) {
            silenceTimeout = t
            i += 1
        }
    case "--locale":
        if i + 1 < args.count {
            locale = args[i + 1]
            i += 1
        }
    default:
        break
    }
    i += 1
}

let dictation = Dictation(locale: locale, silenceTimeout: silenceTimeout)

// Handle SIGTERM (sent by Go when user presses Enter)
signal(SIGTERM) { _ in
    exit(0)
}

signal(SIGINT) { _ in
    exit(0)
}

dictation.start()
RunLoop.main.run()
