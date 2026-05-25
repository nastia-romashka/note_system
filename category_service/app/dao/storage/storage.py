from abc import ABC, abstractmethod


class Storage(ABC):
    __slots__ = []

    @abstractmethod
    def find_one(self, *args, **kwargs):
        raise NotImplementedError

    @abstractmethod
    def find(self, *args, **kwargs):
        raise NotImplementedError

    @abstractmethod
    def create(self, *args, **kwargs):
        raise NotImplementedError

    @abstractmethod
    def update(self, *args, **kwargs):
        raise NotImplementedError

    @abstractmethod
    def delete(self, *args, **kwargs):
        raise NotImplementedError

    @abstractmethod
    def execute(self, *args, **kwargs):
        raise NotImplementedError
