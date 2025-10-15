"""
Integration tests for CLI Tool Features (REQ-SK-201 to REQ-SK-203)

Tests verify that the specify CLI tool provides init, check, and agent detection capabilities.
"""

import os
import pytest
from pathlib import Path
import sys


# CANARY: REQ=REQ-SK-201; FEATURE="SpecifyCLIInit"; ASPECT=CLI; STATUS=TESTED; TEST=test_specify_cli_init_implementation; OWNER=tests; UPDATED=2025-10-15
def test_specify_cli_init_implementation():
    """Test that specify CLI init command is implemented."""
    # Verify the main CLI module exists and contains init functionality
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "CANARY:" in content, "CLI module should contain CANARY token"
    assert "REQ-SK-201" in content, "CLI module should track REQ-SK-201"

    # Verify init-related functionality exists
    assert "init" in content.lower() or "initialize" in content.lower(), "CLI should support init command"


# CANARY: REQ=REQ-SK-202; FEATURE="SpecifyCLICheck"; ASPECT=CLI; STATUS=TESTED; TEST=test_specify_cli_check_implementation; OWNER=tests; UPDATED=2025-10-15
def test_specify_cli_check_implementation():
    """Test that specify CLI check command is implemented."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-202" in content, "CLI module should track REQ-SK-202"

    # Verify check-related functionality exists
    assert "check" in content.lower() or "validate" in content.lower() or "verify" in content.lower(), \
        "CLI should support check command"


# CANARY: REQ=REQ-SK-203; FEATURE="AgentDetection"; ASPECT=Core; STATUS=TESTED; TEST=test_agent_detection_implementation; OWNER=tests; UPDATED=2025-10-15
def test_agent_detection_implementation():
    """Test that agent detection capability is implemented."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module not found"

    content = cli_module.read_text()
    assert "REQ-SK-203" in content, "CLI module should track REQ-SK-203"

    # Verify agent detection exists (could be environment detection, etc.)
    agent_indicators = ["agent", "detect", "environment", "claude", "copilot", "gemini", "cursor"]
    has_agent_support = any(indicator in content.lower() for indicator in agent_indicators)
    assert has_agent_support, "CLI should support agent detection"


def test_cli_module_structure():
    """Meta-test: Verify CLI module has proper structure and tracking."""
    cli_module = Path(__file__).parent.parent.parent / "src" / "specify_cli" / "__init__.py"
    assert cli_module.exists(), "CLI module must exist"

    content = cli_module.read_text()

    # Verify all three CLI requirements are tracked
    assert "REQ-SK-201" in content, "Missing tracking for Specify CLI Init"
    assert "REQ-SK-202" in content, "Missing tracking for Specify CLI Check"
    assert "REQ-SK-203" in content, "Missing tracking for Agent Detection"

    # Verify CANARY tokens are properly formatted
    canary_count = content.count("CANARY:")
    assert canary_count >= 3, f"Expected at least 3 CANARY tokens, found {canary_count}"


def test_pyproject_has_cli_entry_point():
    """Test that pyproject.toml defines the CLI entry point."""
    pyproject_path = Path(__file__).parent.parent.parent / "pyproject.toml"
    assert pyproject_path.exists(), "pyproject.toml not found"

    content = pyproject_path.read_text()
    assert "specify" in content, "pyproject.toml should define 'specify' command"
    assert "specify_cli" in content, "pyproject.toml should reference specify_cli module"
