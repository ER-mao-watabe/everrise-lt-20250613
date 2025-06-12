.PHONY: yolo help

# YOLO mode - Claude Code with dangerous skip permissions
# Usage: make yolo PROMPT="your prompt here"
yolo:
	@if [ -n "$(PROMPT)" ]; then \
		claude --dangerously-skip-permissions "$(PROMPT)"; \
	else \
		claude --dangerously-skip-permissions; \
	fi

# Show help
help:
	@echo "Available commands:"
	@echo "  yolo           - YOLO mode: Start Claude Code with dangerously skip permissions"
	@echo "  yolo PROMPT=\"\" - YOLO mode with a specific prompt"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make yolo"
	@echo "  make yolo PROMPT=\"Let's build something awesome!\""

# Default target
.DEFAULT_GOAL := help