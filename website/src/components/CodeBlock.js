import React from "react";
import Prism from "prismjs";

// Register Glide as a Prism language
Prism.languages.glide = {
  comment: /;.*/,
  string: /"(?:[^"\\]|\\.)*"/,
  number: /\b\d+(?:\.\d+)?\b/,
  keyword:
    /\b(?:fn|let|set|const|if|for|each|loop|recur|return|go|try|guard|use|type|struct|test|assert|bench|in|nil|true|false)\b/,
  builtin: /\b(?:fmt|os|math|strings|sync|errors|sort|strconv|io|bufio)\b/,
  "class-name": /\b[A-Z][a-zA-Z0-9]*\b/,
  operator: /(?:->?>?|<-|[+\-*\/%]|[<>]=?|[!=]=|&&|\|\||!(?!=)|&(?!&))/,
  punctuation: /[()[\]{},.]/,
};

const CodeBlock = ({ code, language = "glide", filename, className: extraClass = "" }) => {
  const lang = Prism.languages[language] ? language : "glide";

  // Prism.highlight returns syntax-coloured HTML wrapping the input code
  // string with <span> elements. The code rendered here is always
  // statically-authored example code, never user-supplied input, so
  // dangerouslySetInnerHTML is safe in this context.
  let highlighted;
  try {
    highlighted = Prism.highlight(
      code,
      Prism.languages[lang] || Prism.languages.glide,
      lang
    );
  } catch {
    highlighted = code
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  return (
    <div className={`rounded-xl overflow-hidden shadow-xl border border-slate-700/50 ${extraClass}`}>
      {filename && (
        <div className="bg-slate-800 border-b border-slate-700 px-4 py-2 flex items-center gap-3">
          <div className="flex gap-1.5">
            <span className="w-3 h-3 rounded-full bg-red-500/70" />
            <span className="w-3 h-3 rounded-full bg-yellow-500/70" />
            <span className="w-3 h-3 rounded-full bg-green-500/70" />
          </div>
          <span className="text-xs text-slate-400 font-mono">{filename}</span>
        </div>
      )}
      <pre className="bg-slate-900 p-5 overflow-x-auto text-sm leading-relaxed m-0">
        <code
          className={`language-${lang}`}
          // eslint-disable-next-line react/no-danger
          dangerouslySetInnerHTML={{ __html: highlighted }}
        />
      </pre>
    </div>
  );
};

export default CodeBlock;
