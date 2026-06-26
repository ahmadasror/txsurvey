import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
import { cn } from "@/lib/utils";
import { typeLabel } from "@/lib/questionTypes";
import type { Question } from "@/types/forms";

interface Props {
  questions: Question[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onReorder: (orderedIds: string[]) => void;
}

export function SortableQuestionList({ questions, selectedId, onSelect, onReorder }: Props) {
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }));

  const onDragEnd = (e: DragEndEvent) => {
    const { active, over } = e;
    if (!over || active.id === over.id) return;
    const oldIndex = questions.findIndex((q) => q.id === active.id);
    const newIndex = questions.findIndex((q) => q.id === over.id);
    onReorder(arrayMove(questions, oldIndex, newIndex).map((q) => q.id));
  };

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={questions.map((q) => q.id)} strategy={verticalListSortingStrategy}>
        <ul className="space-y-1.5">
          {questions.map((q, i) => (
            <Row
              key={q.id}
              question={q}
              index={i}
              selected={q.id === selectedId}
              onSelect={() => onSelect(q.id)}
            />
          ))}
        </ul>
      </SortableContext>
    </DndContext>
  );
}

function Row({
  question,
  index,
  selected,
  onSelect,
}: {
  question: Question;
  index: number;
  selected: boolean;
  onSelect: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: question.id,
  });
  const style = { transform: CSS.Transform.toString(transform), transition };

  return (
    <li
      ref={setNodeRef}
      style={style}
      className={cn(
        "flex items-start gap-2 rounded-xl border bg-card p-2 text-sm transition-colors",
        selected ? "border-primary bg-primary-soft" : "hover:border-primary/40",
        isDragging && "opacity-60",
      )}
    >
      <button
        className="mt-0.5 cursor-grab touch-none text-muted-foreground active:cursor-grabbing"
        {...attributes}
        {...listeners}
        aria-label="Geser untuk mengurutkan"
      >
        <GripVertical className="size-4" />
      </button>
      <span
        className={cn(
          "mt-0.5 grid size-6 shrink-0 place-items-center rounded-lg border text-xs font-semibold tabular-nums",
          selected ? "border-primary bg-primary text-primary-foreground" : "border-input text-muted-foreground",
        )}
      >
        {index + 1}
      </span>
      <button className="flex min-w-0 flex-1 flex-col items-start text-left" onClick={onSelect}>
        <span className="break-words font-medium text-foreground">
          {question.title || <span className="text-muted-foreground">Tanpa judul</span>}
        </span>
        <span className="text-xs text-muted-foreground">
          {typeLabel(question.type)}
          {question.required ? " · wajib" : ""}
        </span>
      </button>
    </li>
  );
}
