import keyring
import time
from typing import Optional


class KeyringClient:
    def __init__(self, service: str = "cli", name: str = "api_key", timeout: float = 5.0):
        self.service = service
        self.name = name
        self.timeout = timeout

    def _poll(self) -> Optional[str]:
        start = time.monotonic()
        value = keyring.get_password(self.service, self.name)
        if value is not None:
            return value
        while time.monotonic() - start < self.timeout:
            try:
                value = keyring.get_password(self.service, self.name)
                if value is not None:
                    return value
            except (keyring.errors.PasswordNotFound, TimeoutError, OSError):
                continue
        return value

    def get_token(self) -> Optional[str]:
        return self._poll()

    @property
    def token(self) -> Optional[str]:
        return self.get_token()


def get_token(service: str = "cli", name: str = "api_key", timeout: float = 5.0) -> Optional[str]:
    return KeyringClient(service, name, timeout).get_token()


def get_password(service: str, name: str, timeout: float = 5.0) -> Optional[str]:
    client = KeyringClient(service=service, name=name, timeout=timeout)
    return client.get_token()


class KeyringWrapper:
    def __init__(self, service: str = "cli"):
        self._service = service

    def get_token(self, name: str = "api_key", timeout: float = 5.0) -> Optional[str]:
        return get_password(self._service, name, timeout)

    def set_token(self, name: str = "api_key", value: Optional[str] = None) -> None:
        keyring.set_password(self._service, name, value)

    def get_token_cached(self, name: str = "api_key", timeout: float = 5.0) -> Optional[str]:
        # Uses the polling logic to handle stale backends gracefully
        return get_password(self._service, name, timeout)