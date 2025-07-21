# Contributing to This Project

Thank you for your interest in contributing to our project! We welcome contributions from everyone. This document provides guidelines and steps for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/IdanKoblik/OpenPipe.git
   ```
3. Add the original repository as upstream:
   ```bash
   git remote add upstream https://github.com/IdanKoblik/OpenPipe.git
   ```

## Development Process

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
   or
   ```bash
   git checkout -b fix/your-bugfix-name
   ```

2. Make your changes and commit them:
   ```bash
   git add .
   git commit -m "Description of your changes"
   ```

3. Keep your branch updated with upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

## Pull Request Process

1. Push your changes to your fork:
   ```bash
   git push origin your-branch-name
   ```

2. Go to the original repository and create a Pull Request (PR)
3. Fill in the PR template with all required information
4. Wait for review and address any feedback

## Code Style Guidelines

- Write clear, readable, and maintainable code
- Follow the existing code style of the project
- Include comments where necessary
- Write meaningful commit messages
- Update documentation as needed

## Testing

- Add tests for new features
- Ensure all tests pass before submitting your PR
- Update existing tests if necessary

## Reporting Issues

- Use the GitHub issue tracker
- Check if the issue already exists before creating a new one
- Include as much detail as possible
- Follow the issue template if provided

