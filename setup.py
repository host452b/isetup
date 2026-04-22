"""
setuptools driver. All metadata lives in pyproject.toml; this file exists only
to override bdist_wheel so each wheel is tagged for a specific platform.

Each platform wheel embeds the matching pre-compiled Go binary under
isetup/bin/ — we flag the wheel as impure and force a py3-none-<plat> tag so
pip installs the right archive on the user's machine.
"""

from setuptools import setup
from setuptools.command.bdist_wheel import bdist_wheel


class PlatformSpecificWheel(bdist_wheel):
    def finalize_options(self):
        super().finalize_options()
        self.root_is_pure = False

    def get_tag(self):
        _, _, plat = super().get_tag()
        return "py3", "none", plat


setup(cmdclass={"bdist_wheel": PlatformSpecificWheel})
