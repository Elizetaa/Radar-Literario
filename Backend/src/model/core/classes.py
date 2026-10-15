from sqlalchemy import String, create_engine, select
from sqlalchemy.orm import DeclarativeBase, Mapped, Session, mapped_column

class Base(DeclarativeBase):
    pass

class Leitor(Base):
    __tablename__ = "readers"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(50))
    user_name: Mapped[str] = mapped_column(String(50))
    email: Mapped[str] = mapped_column(String(50))
    password: Mapped[str] = mapped_column(String(50))
    born_date: Mapped[str] = mapped_column(String(50))
    gender: Mapped[str] = mapped_column(String(50))
    
    