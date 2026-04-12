from pydantic import BaseModel, ConfigDict


class Base(BaseModel):
    model_config = ConfigDict(
        extra="forbid",
        from_attributes=True,
        populate_by_name=True,
    )
