package middleware

import "net/http"

// SecurityHeaders applies modern, OWASP-compliant HTTP security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Clickjacking protection (legacy header fallback for older browsers)
		w.Header().Set("X-Frame-Options", "DENY")

		// Control referrer exposure
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable unused browser hardware APIs
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Process isolation & cross-origin policies
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		// Note on X-XSS-Protection:
		// Intentionally omitted. OWASP and MDN specify X-XSS-Protection as deprecated and obsolete.
		// In modern browsers, it can introduce security vulnerabilities. XSS defense is handled via CSP.

		// Modern Content Security Policy (CSP) compatible with Markdown, KaTeX, Web Workers & Monaco Editor
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: blob: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' https:; " +
			"worker-src 'self' blob:; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none';"
		w.Header().Set("Content-Security-Policy", csp)

		// Strict Transport Security (HSTS) when running over HTTPS / secure proxy
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}
