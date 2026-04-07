import React, { useState } from "react";
import { Link } from "gatsby";

const NAV_LINKS = [
  { label: "Guide", to: "/guide" },
  { label: "Examples", to: "/examples" },
  {
    label: "GitHub",
    href: "https://github.com/elielamora/glide",
    external: true,
  },
];

const Navbar = () => {
  const [menuOpen, setMenuOpen] = useState(false);
  return (
    <nav className="sticky top-0 z-50 bg-white/90 backdrop-blur border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <Link
          to="/"
          className="flex items-center gap-2 font-bold text-xl text-glide-700 hover:text-glide-900 transition-colors"
        >
          <svg
            viewBox="0 0 32 32"
            width="28"
            height="28"
            fill="none"
            aria-hidden="true"
          >
            <rect width="32" height="32" rx="8" fill="#7c3aed" />
            <text
              x="7"
              y="23"
              fontFamily="monospace"
              fontSize="18"
              fontWeight="bold"
              fill="white"
            >
              G
            </text>
          </svg>
          glide
        </Link>

        <ul className="hidden sm:flex items-center gap-6">
          {NAV_LINKS.map(({ label, to, href, external }) => (
            <li key={label}>
              {external ? (
                <a
                  href={href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-gray-600 hover:text-glide-700 font-medium transition-colors"
                >
                  {label}
                </a>
              ) : (
                <Link
                  to={to}
                  className="text-gray-600 hover:text-glide-700 font-medium transition-colors"
                  activeClassName="text-glide-700"
                >
                  {label}
                </Link>
              )}
            </li>
          ))}
          <li>
            <a
              href="https://github.com/elielamora/glide"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-glide-600 text-white text-sm font-semibold hover:bg-glide-700 transition-colors"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="currentColor"
                aria-hidden="true"
              >
                <path d="M12 2C6.477 2 2 6.477 2 12c0 4.418 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.009-.868-.013-1.703-2.782.604-3.369-1.342-3.369-1.342-.454-1.155-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0 1 12 6.836a9.59 9.59 0 0 1 2.504.337c1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.163 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
              </svg>
              Star on GitHub
            </a>
          </li>
        </ul>

        {/* Mobile menu toggle */}
        <button
          className="sm:hidden p-2 rounded-md text-gray-500 hover:text-glide-700 hover:bg-glide-50 transition-colors"
          aria-label={menuOpen ? "Close menu" : "Open menu"}
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen((v) => !v)}
        >
          {menuOpen ? (
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          ) : (
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          )}
        </button>
      </div>

      {/* Mobile dropdown */}
      {menuOpen && (
        <div className="sm:hidden border-t border-gray-100 bg-white/95 backdrop-blur">
          <ul className="px-4 py-3 space-y-1">
            {NAV_LINKS.map(({ label, to, href, external }) => (
              <li key={label}>
                {external ? (
                  <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block px-3 py-2 rounded-md text-gray-700 hover:text-glide-700 hover:bg-glide-50 font-medium transition-colors"
                    onClick={() => setMenuOpen(false)}
                  >
                    {label}
                  </a>
                ) : (
                  <Link
                    to={to}
                    className="block px-3 py-2 rounded-md text-gray-700 hover:text-glide-700 hover:bg-glide-50 font-medium transition-colors"
                    activeClassName="text-glide-700 bg-glide-50"
                    onClick={() => setMenuOpen(false)}
                  >
                    {label}
                  </Link>
                )}
              </li>
            ))}
            <li className="pt-1">
              <a
                href="https://github.com/elielamora/glide"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-3 py-2 rounded-md bg-glide-600 text-white font-semibold text-sm hover:bg-glide-700 transition-colors"
                onClick={() => setMenuOpen(false)}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path d="M12 2C6.477 2 2 6.477 2 12c0 4.418 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.009-.868-.013-1.703-2.782.604-3.369-1.342-3.369-1.342-.454-1.155-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0 1 12 6.836a9.59 9.59 0 0 1 2.504.337c1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.163 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
                </svg>
                Star on GitHub
              </a>
            </li>
          </ul>
        </div>
      )}
    </nav>
  );
};

const Footer = () => (
  <footer className="bg-gray-950 text-gray-400 py-12 mt-auto">
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-8 mb-8">
        <div>
          <p className="font-bold text-white mb-2 text-lg">glide</p>
          <p className="text-sm">
            The readable Lisp that compiles to Go. Open source, MIT licensed.
          </p>
        </div>
        <div>
          <p className="font-semibold text-gray-300 mb-2">Documentation</p>
          <ul className="space-y-1 text-sm">
            <li>
              <Link to="/guide" className="hover:text-white transition-colors">
                Developer Guide
              </Link>
            </li>
            <li>
              <Link
                to="/examples"
                className="hover:text-white transition-colors"
              >
                Examples
              </Link>
            </li>
          </ul>
        </div>
        <div>
          <p className="font-semibold text-gray-300 mb-2">Community</p>
          <ul className="space-y-1 text-sm">
            <li>
              <a
                href="https://github.com/elielamora/glide"
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-white transition-colors"
              >
                GitHub
              </a>
            </li>
            <li>
              <a
                href="https://github.com/elielamora/glide/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-white transition-colors"
              >
                Issues
              </a>
            </li>
          </ul>
        </div>
      </div>
      <div className="border-t border-gray-800 pt-6 text-sm text-center">
        © {new Date().getFullYear()} Glide Contributors. MIT License.
      </div>
    </div>
  </footer>
);

const Layout = ({ children }) => (
  <div className="min-h-screen flex flex-col font-sans antialiased">
    <Navbar />
    <main className="flex-1">{children}</main>
    <Footer />
  </div>
);

export default Layout;
