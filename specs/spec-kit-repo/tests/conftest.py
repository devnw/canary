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
Pytest configuration for spec-kit test suite.

This file is automatically loaded by pytest and provides shared fixtures
and configuration for all tests.
"""

import sys
from pathlib import Path

# Add src directory to Python path for importing specify_cli
spec_kit_root = Path(__file__).parent.parent
src_path = spec_kit_root / "src"
if str(src_path) not in sys.path:
    sys.path.insert(0, str(src_path))


def pytest_configure(config):
    """Configure pytest with custom markers."""
    config.addinivalue_line(
        "markers", "integration: mark test as an integration test"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow running"
    )
