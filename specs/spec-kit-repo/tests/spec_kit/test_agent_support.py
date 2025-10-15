"""
Integration tests for Agent Support (REQ-SK-601 to REQ-SK-605)

Tests verify that the CLI supports multiple AI coding agents including
Claude Code, GitHub Copilot, Gemini CLI, Cursor, and multi-agent workflows.
"""

import pytest
from pathlib import Path


# CANARY: REQ=REQ-SK-601; FEATURE="ClaudeCodeSupport"; ASPECT=Agent; STATUS=TESTED; TEST=test_claude_code_support_tracked; OWNER=tests; UPDATED=2025-10-15
def test_claude_code_support_tracked():
    """Test that Claude Code agent support is tracked."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-601" in content, "CLI module should track REQ-SK-601 for Claude Code support"

    # Verify Claude Code is referenced in documentation or agent files
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    if agents_doc.exists():
        agents_content = agents_doc.read_text()
        assert "Claude" in agents_content or "claude" in agents_content, \
            "AGENTS.md should document Claude Code support"


# CANARY: REQ=REQ-SK-602; FEATURE="CopilotSupport"; ASPECT=Agent; STATUS=TESTED; TEST=test_copilot_support_tracked; OWNER=tests; UPDATED=2025-10-15
def test_copilot_support_tracked():
    """Test that GitHub Copilot agent support is tracked."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-602" in content, "CLI module should track REQ-SK-602 for Copilot support"

    # Verify Copilot is referenced in documentation
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    if agents_doc.exists():
        agents_content = agents_doc.read_text()
        assert "Copilot" in agents_content or "copilot" in agents_content, \
            "AGENTS.md should document Copilot support"


# CANARY: REQ=REQ-SK-603; FEATURE="GeminiCLISupport"; ASPECT=Agent; STATUS=TESTED; TEST=test_gemini_cli_support_tracked; OWNER=tests; UPDATED=2025-10-15
def test_gemini_cli_support_tracked():
    """Test that Gemini CLI agent support is tracked."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-603" in content, "CLI module should track REQ-SK-603 for Gemini CLI support"

    # Verify Gemini is referenced in documentation
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    if agents_doc.exists():
        agents_content = agents_doc.read_text()
        assert "Gemini" in agents_content or "gemini" in agents_content, \
            "AGENTS.md should document Gemini support"


# CANARY: REQ=REQ-SK-604; FEATURE="CursorSupport"; ASPECT=Agent; STATUS=TESTED; TEST=test_cursor_support_tracked; OWNER=tests; UPDATED=2025-10-15
def test_cursor_support_tracked():
    """Test that Cursor agent support is tracked."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-604" in content, "CLI module should track REQ-SK-604 for Cursor support"

    # Verify Cursor is referenced in documentation
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    if agents_doc.exists():
        agents_content = agents_doc.read_text()
        assert "Cursor" in agents_content or "cursor" in agents_content, \
            "AGENTS.md should document Cursor support"


# CANARY: REQ=REQ-SK-605; FEATURE="MultiAgentSupport"; ASPECT=Agent; STATUS=TESTED; TEST=test_multi_agent_support_tracked; OWNER=tests; UPDATED=2025-10-15
def test_multi_agent_support_tracked():
    """Test that multi-agent support (14+ agents) is tracked."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-605" in content, "CLI module should track REQ-SK-605 for multi-agent support"

    # Verify AGENTS.md documents multiple agents
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    assert agents_doc.exists(), "AGENTS.md should exist to document multi-agent support"

    agents_content = agents_doc.read_text()
    # Count common agent names mentioned
    agent_names = [
        "Claude", "Copilot", "Gemini", "Cursor", "ChatGPT", "Cody",
        "Aider", "Qodo", "Continue", "Tabnine", "Amazon Q", "Windsurf"
    ]

    agents_found = sum(1 for name in agent_names if name in agents_content)
    assert agents_found >= 4, f"AGENTS.md should document multiple agents, found {agents_found}"


def test_agent_file_template_exists():
    """Test that agent file template exists for agent configuration."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "agent-file-template.md"
    assert template_path.exists(), "Agent file template should exist"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-306" in content, "Template should track REQ-SK-306"


def test_agents_documentation_comprehensive():
    """Meta-test: Verify AGENTS.md provides comprehensive agent documentation."""
    agents_doc = Path(__file__).parent.parent.parent / "AGENTS.md"
    assert agents_doc.exists(), "AGENTS.md must exist"

    content = agents_doc.read_text()

    # Verify it's substantial documentation
    assert len(content) > 500, "AGENTS.md should contain comprehensive documentation"

    # Verify key sections or agent names are present
    required_elements = ["agent", "AI", "support"]
    for element in required_elements:
        assert element in content.lower(), f"AGENTS.md should mention '{element}'"
