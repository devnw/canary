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
Integration tests for Template System (REQ-SK-301 to REQ-SK-306)

Tests verify that all template files exist, have proper structure, and are properly tracked.
"""

import pytest
from pathlib import Path


# CANARY: REQ=REQ-SK-301; FEATURE="SpecTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_spec_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_spec_template_exists():
    """Test that spec template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "spec-template.md"
    assert template_path.exists(), f"Spec template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-301" in content, "Template should track REQ-SK-301"

    # Verify template contains key sections
    assert len(content) > 500, "Template should contain substantial content"

    # Check for template structure indicators
    has_structure = any(indicator in content.lower() for indicator in [
        "spec", "feature", "requirement", "description", "objective"
    ])
    assert has_structure, "Template should contain specification structure"


# CANARY: REQ=REQ-SK-302; FEATURE="PlanTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_plan_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_plan_template_exists():
    """Test that plan template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "plan-template.md"
    assert template_path.exists(), f"Plan template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-302" in content, "Template should track REQ-SK-302"

    # Verify template contains substantial content
    assert len(content) > 500, "Template should contain substantial content"

    # Check for plan-related content
    has_plan_content = any(indicator in content.lower() for indicator in [
        "plan", "design", "architecture", "approach", "strategy"
    ])
    assert has_plan_content, "Template should contain planning structure"


# CANARY: REQ=REQ-SK-303; FEATURE="TasksTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_tasks_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_tasks_template_exists():
    """Test that tasks template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "tasks-template.md"
    assert template_path.exists(), f"Tasks template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-303" in content, "Template should track REQ-SK-303"

    # Verify template contains substantial content
    assert len(content) > 200, "Template should contain substantial content"

    # Check for task-related content
    has_task_content = any(indicator in content.lower() for indicator in [
        "task", "todo", "action", "step", "checklist"
    ])
    assert has_task_content, "Template should contain task structure"


# CANARY: REQ=REQ-SK-304; FEATURE="ChecklistTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_checklist_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_checklist_template_exists():
    """Test that checklist template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "checklist-template.md"
    assert template_path.exists(), f"Checklist template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-304" in content, "Template should track REQ-SK-304"

    # Verify template contains substantial content
    assert len(content) > 200, "Template should contain substantial content"

    # Check for checklist-related content
    has_checklist_content = any(indicator in content.lower() for indicator in [
        "checklist", "check", "verify", "validate", "review"
    ])
    assert has_checklist_content, "Template should contain checklist structure"


# CANARY: REQ=REQ-SK-305; FEATURE="ConstitutionTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_constitution_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_constitution_template_exists():
    """Test that constitution template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "memory" / "constitution.md"
    assert template_path.exists(), f"Constitution template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-305" in content, "Template should track REQ-SK-305"

    # Verify template contains substantial content (constitution should be detailed)
    assert len(content) > 1000, "Constitution should contain substantial content"

    # Check for constitution-related content
    has_constitution_content = any(indicator in content.lower() for indicator in [
        "constitution", "principle", "article", "mandate", "governance"
    ])
    assert has_constitution_content, "Template should contain constitutional structure"


# CANARY: REQ=REQ-SK-306; FEATURE="AgentFileTemplate"; ASPECT=Templates; STATUS=TESTED; TEST=test_agent_file_template_exists; OWNER=tests; UPDATED=2025-10-15
def test_agent_file_template_exists():
    """Test that agent file template exists and is valid."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "agent-file-template.md"
    assert template_path.exists(), f"Agent file template not found at {template_path}"

    content = template_path.read_text()
    assert "CANARY:" in content, "Template should contain CANARY token"
    assert "REQ-SK-306" in content, "Template should track REQ-SK-306"

    # Verify template contains substantial content
    assert len(content) > 200, "Template should contain substantial content"

    # Check for agent-related content
    has_agent_content = any(indicator in content.lower() for indicator in [
        "agent", "context", "ai", "assistant", "guidance"
    ])
    assert has_agent_content, "Template should contain agent configuration structure"


def test_all_templates_tracked():
    """Meta-test: Verify all template files are tracked with CANARY tokens."""
    templates_dir = Path(__file__).parent.parent.parent / "templates"
    assert templates_dir.exists(), "Templates directory should exist"

    expected_templates = [
        "spec-template.md",
        "plan-template.md",
        "tasks-template.md",
        "checklist-template.md",
        "agent-file-template.md"
    ]

    for template_name in expected_templates:
        template_path = templates_dir / template_name
        assert template_path.exists(), f"Template {template_name} not found"

        content = template_path.read_text()
        assert "CANARY:" in content, f"Template {template_name} missing CANARY token"


def test_constitution_special_location():
    """Meta-test: Verify constitution is in special memory/ directory."""
    constitution_path = Path(__file__).parent.parent.parent / "memory" / "constitution.md"
    assert constitution_path.exists(), "Constitution should exist in memory/ directory"

    content = constitution_path.read_text()
    assert "CANARY:" in content, "Constitution missing CANARY token"

    # Verify multiple constitutional requirements are tracked
    canary_count = content.count("CANARY:")
    assert canary_count >= 6, f"Constitution should track multiple requirements, found {canary_count}"
