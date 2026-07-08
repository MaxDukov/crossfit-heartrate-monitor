"""CRUD API для управления инвентарём зала."""

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Equipment, GymInventory
from ..schemas import EquipmentOut, GymInventoryOut, GymInventoryUpdate

router = APIRouter(prefix="/api/equipment", tags=["equipment"])


@router.get("", response_model=list[EquipmentOut])
def list_equipment(db: Session = Depends(get_db)):
    """Возвращает полный каталог инвентаря."""
    return db.query(Equipment).order_by(Equipment.category, Equipment.name).all()


@router.get("/inventory", response_model=list[GymInventoryOut])
def list_inventory(db: Session = Depends(get_db)):
    """Возвращает инвентарь, доступный в зале."""
    return db.query(GymInventory).all()


@router.put("/inventory", response_model=list[GymInventoryOut])
def update_inventory(data: GymInventoryUpdate, db: Session = Depends(get_db)):
    """Полностью перезаписывает инвентарь зала."""
    db.query(GymInventory).delete()
    for item in data.items:
        db.add(GymInventory(
            equipment_key=item["equipment_key"],
            quantity=item.get("quantity", 1),
        ))
    db.commit()
    return db.query(GymInventory).all()
