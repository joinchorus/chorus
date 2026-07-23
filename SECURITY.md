# Security Policy

## Official Domains & Security

Chorus operates strictly across standardized public subdomains:
- **Public Site:** [joinchorus.app](https://joinchorus.app)
- **Application & API:** [chat.joinchorus.app](https://chat.joinchorus.app)
- **Documentation & Security Guides:** [docs.joinchorus.app/security](https://docs.joinchorus.app/security)

---

## Reporting Vulnerabilities

We take the security and anonymity guarantees of Chorus seriously. If you discover a security vulnerability or potential privacy leak, please report it to us confidentially before publishing it publicly.

### How to Report

- **Email**: Send details to `barissalih@babacan.me` or contact the maintainers directly.
- **GitHub Private Vulnerability Reporting**: Use the **Security** tab on the [Chorus GitHub Repository](https://github.com/barissalihbabacan/Chorus/security) to submit a confidential report.
- **Documentation & Security Policy**: Available at [docs.joinchorus.app/security](https://docs.joinchorus.app/security).

Please include:
1. A description of the vulnerability and its potential impact.
2. Step-by-step reproduction instructions or a minimal proof-of-concept.
3. Any proposed mitigations or fixes.

---

## Security Scope & Core Guarantees

We are particularly interested in reports related to:
- **Anonymity Breaches**: Vulnerabilities that allow cross-thread correlation or unmasking of ephemeral identities.
- **IP Address Leaks**: Exposure of client IP addresses beyond the server's coarse geolocation handling.
- **Data Store Injection**: Arbitrary file execution or traversal via NDJSON or JSON storage paths.
- **HTTP Handler Vulnerabilities**: Unhandled panics, resource exhaustion, or payload injection.

---

## Disclosure Process

1. **Acknowledgment**: We will acknowledge receipt of your security report within **48 hours**.
2. **Assessment**: The maintainers will investigate and determine the severity of the issue.
3. **Patching**: We will prepare a fix and release a security patch (`v0.x.y`).
4. **Public Disclosure**: Once the fix is released, we will publish a security advisory giving full credit to the reporter.

Thank you for helping keep Chorus secure and private for everyone.
