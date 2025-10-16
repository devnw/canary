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
Integration tests for Documentation System (REQ-SK-701 to REQ-SK-704)

Tests verify that all documentation files exist, have proper structure, and comprehensive content.
"""

import pytest
from pathlib import Path


# CANARY: REQ=REQ-SK-701; FEATURE="QuickstartGuide"; ASPECT=Documentation; STATUS=TESTED; TEST=test_quickstart_guide_exists; OWNER=tests; UPDATED=2025-10-15
def test_quickstart_guide_exists():
    """Test that quickstart guide exists and is comprehensive."""
    quickstart_path = Path(__file__).parent.parent.parent / "docs" / "quickstart.md"
    assert quickstart_path.exists(), f"Quickstart guide not found at {quickstart_path}"

    content = quickstart_path.read_text()
    assert "CANARY:" in content, "Quickstart guide should contain CANARY token"
    assert "REQ-SK-701" in content, "Quickstart guide should track REQ-SK-701"

    # Verify substantial content for a quickstart guide
    assert len(content) > 500, "Quickstart guide should contain substantial content"

    # Check for quickstart elements
    quickstart_indicators = [
        "quick", "start", "install", "setup", "getting started", "introduction"
    ]
    has_quickstart_content = any(indicator in content.lower() for indicator in quickstart_indicators)
    assert has_quickstart_content, "Guide should contain quickstart content"


# CANARY: REQ=REQ-SK-702; FEATURE="ResearchDocumentation"; ASPECT=Documentation; STATUS=TESTED; TEST=test_research_documentation_exists; OWNER=tests; UPDATED=2025-10-15
def test_research_documentation_exists():
    """Test that research documentation exists in index."""
    docs_index_path = Path(__file__).parent.parent.parent / "docs" / "index.md"
    assert docs_index_path.exists(), f"Documentation index not found at {docs_index_path}"

    content = docs_index_path.read_text()
    assert "CANARY:" in content, "Documentation should contain CANARY token"
    assert "REQ-SK-702" in content, "Documentation should track REQ-SK-702"

    # Verify substantial research documentation content
    assert len(content) > 1000, "Research documentation should be comprehensive"

    # Check for research/methodology content
    research_indicators = [
        "research", "methodology", "approach", "rationale", "design decision"
    ]
    has_research_content = any(indicator in content.lower() for indicator in research_indicators)
    assert has_research_content, "Documentation should contain research content"


# CANARY: REQ=REQ-SK-703; FEATURE="DataModelDocumentation"; ASPECT=Documentation; STATUS=TESTED; TEST=test_data_model_documentation_exists; OWNER=tests; UPDATED=2025-10-15
def test_data_model_documentation_exists():
    """Test that data model documentation exists in index."""
    docs_index_path = Path(__file__).parent.parent.parent / "docs" / "index.md"
    assert docs_index_path.exists(), f"Documentation index not found at {docs_index_path}"

    content = docs_index_path.read_text()
    assert "REQ-SK-703" in content, "Documentation should track REQ-SK-703"

    # Check for data model content
    data_model_indicators = [
        "data", "model", "schema", "structure", "entity", "field", "format"
    ]
    has_data_model_content = any(indicator in content.lower() for indicator in data_model_indicators)
    assert has_data_model_content, "Documentation should contain data model content"


# CANARY: REQ=REQ-SK-704; FEATURE="APIContractDocumentation"; ASPECT=Documentation; STATUS=TESTED; TEST=test_api_contract_documentation_exists; OWNER=tests; UPDATED=2025-10-15
def test_api_contract_documentation_exists():
    """Test that API contract documentation exists in index."""
    docs_index_path = Path(__file__).parent.parent.parent / "docs" / "index.md"
    assert docs_index_path.exists(), f"Documentation index not found at {docs_index_path}"

    content = docs_index_path.read_text()
    assert "REQ-SK-704" in content, "Documentation should track REQ-SK-704"

    # Check for API/contract content
    api_indicators = [
        "api", "interface", "contract", "endpoint", "method", "function", "command"
    ]
    has_api_content = any(indicator in content.lower() for indicator in api_indicators)
    assert has_api_content, "Documentation should contain API contract content"


def test_docs_directory_structure():
    """Meta-test: Verify documentation directory has proper structure."""
    docs_dir = Path(__file__).parent.parent.parent / "docs"
    assert docs_dir.exists(), "Documentation directory should exist"

    # Verify key documentation files exist
    assert (docs_dir / "index.md").exists(), "index.md should exist"
    assert (docs_dir / "quickstart.md").exists(), "quickstart.md should exist"


def test_all_documentation_tracked():
    """Meta-test: Verify all documentation files contain CANARY tokens."""
    docs_dir = Path(__file__).parent.parent.parent / "docs"

    required_docs = ["index.md", "quickstart.md"]

    for doc_name in required_docs:
        doc_path = docs_dir / doc_name
        assert doc_path.exists(), f"Documentation {doc_name} not found"

        content = doc_path.read_text()
        assert "CANARY:" in content, f"Documentation {doc_name} missing CANARY token"


def test_index_consolidates_multiple_requirements():
    """Meta-test: Verify index.md consolidates research, data model, and API docs."""
    docs_index_path = Path(__file__).parent.parent.parent / "docs" / "index.md"
    assert docs_index_path.exists(), "Documentation index should exist"

    content = docs_index_path.read_text()

    # Verify all three consolidated requirements are tracked
    assert "REQ-SK-702" in content, "Index should track research documentation"
    assert "REQ-SK-703" in content, "Index should track data model documentation"
    assert "REQ-SK-704" in content, "Index should track API contract documentation"

    # Count CANARY tokens for documentation requirements
    canary_count = content.count("REQ-SK-70")
    assert canary_count >= 3, f"Index should track multiple doc requirements, found {canary_count}"
