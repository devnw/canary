# Copyright (c) 2024 by CodePros.
#
# This software is proprietary information of CodePros.
# Unauthorized use, copying, modification, distribution, and/or
# disclosure is strictly prohibited, except as provided under the terms
# of the commercial license agreement you have entered into with
# CodePros.
#
# For more details, see the LICENSE file in the root directory of this
# source code repository or contact CodePros at info@codepros.org.

"""
Integration tests for Core Workflow Commands (REQ-SK-101 to REQ-SK-108)

Tests verify that all /speckit.* commands are properly configured and accessible.
"""

import os
import pytest
from pathlib import Path


# CANARY: REQ=REQ-SK-101; FEATURE="ConstitutionCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_constitution_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_constitution_command_exists():
    """Test that constitution command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "constitution.md"
    assert template_path.exists(), f"Constitution command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-101" in content, "Template should track REQ-SK-101"
    assert len(content) > 100, "Template should contain substantial content"


# CANARY: REQ=REQ-SK-102; FEATURE="SpecifyCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_specify_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_specify_command_exists():
    """Test that specify command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "specify.md"
    assert template_path.exists(), f"Specify command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-102" in content, "Template should track REQ-SK-102"
    assert "spec.md" in content or "specification" in content.lower(), "Template should reference specification generation"


# CANARY: REQ=REQ-SK-103; FEATURE="ClarifyCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_clarify_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_clarify_command_exists():
    """Test that clarify command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "clarify.md"
    assert template_path.exists(), f"Clarify command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-103" in content, "Template should track REQ-SK-103"
    assert "ambiguity" in content.lower() or "clarify" in content.lower(), "Template should reference clarification"


# CANARY: REQ=REQ-SK-104; FEATURE="PlanCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_plan_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_plan_command_exists():
    """Test that plan command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "plan.md"
    assert template_path.exists(), f"Plan command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-104" in content, "Template should track REQ-SK-104"
    assert "plan.md" in content or "planning" in content.lower(), "Template should reference plan generation"


# CANARY: REQ=REQ-SK-105; FEATURE="TasksCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_tasks_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_tasks_command_exists():
    """Test that tasks command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "tasks.md"
    assert template_path.exists(), f"Tasks command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-105" in content, "Template should track REQ-SK-105"
    assert "tasks.md" in content or "task" in content.lower(), "Template should reference task management"


# CANARY: REQ=REQ-SK-106; FEATURE="ImplementCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_implement_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_implement_command_exists():
    """Test that implement command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "implement.md"
    assert template_path.exists(), f"Implement command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-106" in content, "Template should track REQ-SK-106"
    assert "implement" in content.lower(), "Template should reference implementation"


# CANARY: REQ=REQ-SK-107; FEATURE="AnalyzeCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_analyze_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_analyze_command_exists():
    """Test that analyze command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "analyze.md"
    assert template_path.exists(), f"Analyze command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-107" in content, "Template should track REQ-SK-107"
    assert "analyze" in content.lower() or "analysis" in content.lower(), "Template should reference analysis"


# CANARY: REQ=REQ-SK-108; FEATURE="ChecklistCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_checklist_command_exists; OWNER=tests; UPDATED=2025-10-15
def test_checklist_command_exists():
    """Test that checklist command template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "checklist.md"
    assert template_path.exists(), f"Checklist command template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-108" in content, "Template should track REQ-SK-108"
    assert "checklist" in content.lower(), "Template should reference checklist functionality"


def test_all_workflow_commands_tracked():
    """Meta-test: Verify all 8 workflow commands are tracked with CANARY tokens."""
    commands_dir = Path(__file__).parent.parent.parent / "templates" / "commands"
    expected_commands = [
        "constitution.md",
        "specify.md",
        "clarify.md",
        "plan.md",
        "tasks.md",
        "implement.md",
        "analyze.md",
        "checklist.md"
    ]

    for cmd_file in expected_commands:
        cmd_path = commands_dir / cmd_file
        assert cmd_path.exists(), f"Command template {cmd_file} not found"

        content = cmd_path.read_text()
        assert "CANARY:" in content, f"Command {cmd_file} missing CANARY token"
