# Contributing to CrossFit Heart Rate Monitor

Thank you for your interest in contributing to the CrossFit Heart Rate Monitor project! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in the Issues section
2. Use the bug report template when creating a new issue
3. Include detailed steps to reproduce the bug
4. Add screenshots or logs if applicable
5. Specify your environment (OS, Python version, etc.)

### Suggesting Features

1. Check if the feature has already been suggested
2. Use the feature request template
3. Provide a clear description of the feature
4. Explain why this feature would be useful
5. Include any relevant examples or mockups

### Pull Requests

1. Fork the repository
2. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes
4. Run tests if applicable
5. Update documentation if needed
6. Commit your changes with clear commit messages
7. Push to your fork
8. Create a Pull Request

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/MaxDukov/crossfit-heartrate-monitor.git
   cd crossfit-hr-monitor
   ```

2. Create and activate a virtual environment:
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

3. Install development dependencies:
   ```bash
   pip install -r requirements-dev.txt
   ```

4. Install pre-commit hooks:
   ```bash
   pre-commit install
   ```

### Code Style

- Follow PEP 8 guidelines
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused
- Write docstrings for all functions and classes

### Testing

- Write tests for new features
- Ensure all tests pass before submitting a PR
- Update tests when modifying existing features
- Aim for good test coverage

### Documentation

- Update README.md if needed
- Add docstrings to new functions and classes
- Update API documentation if applicable
- Include examples for new features

## Project Structure

```
crossfit-hr-monitor/
├── hr.py                 # Main application file
├── requirements.txt      # Production dependencies
├── requirements-dev.txt  # Development dependencies
├── templates/           # HTML templates
│   ├── index.html
│   └── athletes.html
├── static/             # Static files
│   ├── css/
│   └── js/
└── tests/             # Test files
```

## Development Workflow

1. Create an issue for your feature/bug
2. Get approval from maintainers
3. Create a feature branch
4. Make your changes
5. Write/update tests
6. Update documentation
7. Submit a PR
8. Address review comments
9. Get approval and merge

## Commit Messages

Follow these guidelines for commit messages:

- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally
- Consider starting the commit message with an applicable emoji:
  - 🎨 `:art:` when improving the format/structure of the code
  - 🐎 `:racehorse:` when improving performance
  - 🚱 `:non-potable_water:` when plugging memory leaks
  - 📝 `:memo:` when writing docs
  - 🐛 `:bug:` when fixing a bug
  - 🔥 `:fire:` when removing code or files
  - 💚 `:green_heart:` when fixing the CI build
  - ✅ `:white_check_mark:` when adding tests
  - 🔒 `:lock:` when dealing with security
  - ⬆️ `:arrow_up:` when upgrading dependencies
  - ⬇️ `:arrow_down:` when downgrading dependencies

## Review Process

1. All PRs require at least one review
2. Maintainers will review your code
3. Address any feedback
4. Once approved, your PR will be merged

## Questions?

Feel free to open an issue for any questions about contributing.

Thank you for contributing to CrossFit Heart Rate Monitor! 