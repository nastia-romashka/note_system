import { useEffect, useMemo, useRef, useState } from "react";
import {
  Background,
  Controls,
  Handle,
  MarkerType,
  MiniMap,
  Position,
  ReactFlow,
  useEdgesState,
  useNodesState,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import WorkspaceContextMenu from "../../components/WorkspaceContextMenu";
import { buildCategoryOptions, filterGraph } from "./graphFilters";

const NODE_WIDTH = 220;
const CATEGORY_X = 40;
const CATEGORY_DEPTH_GAP = 280;
const NOTE_GAP_X = 350;
const ROW_GAP = 130;
const NODE_APPEAR_DELAY_MS = 500;

const DEFAULT_FILTERS = {
  showCategories: true,
  showNotes: true,
  selectedCategoryId: "",
};

const nodeTypes = {
  graphNode: GraphNode,
};

export default function GraphPage({
  currentWorkspace,
  workspaces,
  graph,
  loading,
  onCreateGraphLink,
  onDeleteGraphLink,
  onBackToNotes,
  onOpenCalendar,
  onOpenGraphNode,
  onOpenProfile,
  onSelectPersonalWorkspace,
  onSelectWorkspace,
  onCreateWorkspace,
}) {
  const [filters, setFilters] = useState(DEFAULT_FILTERS);
  const isWorkspaceMode = Boolean(currentWorkspace);
  const pageTitle = isWorkspaceMode ? `${currentWorkspace.name}: Граф` : "Граф";
  const contextButtonLabel = isWorkspaceMode ? "Настройки пространства" : "Личный кабинет";
  const filteredGraph = useMemo(() => filterGraph(graph, filters), [graph, filters]);
  const preparedGraph = useMemo(() => toFlowGraph(filteredGraph), [filteredGraph]);
  const categoryOptions = useMemo(() => buildCategoryOptions(graph), [graph]);
  const animationOrder = useMemo(
    () => [...preparedGraph.nodes].sort(compareAnimationNodes),
    [preparedGraph.nodes],
  );
  const timerRef = useRef(null);
  const [isAnimating, setIsAnimating] = useState(false);
  const [animationStep, setAnimationStep] = useState(-1);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  const hasActiveFilters =
    filters.selectedCategoryId !== "" || !filters.showCategories || !filters.showNotes;

  const visibleNodeIds = useMemo(() => {
    if (!isAnimating) {
      return new Set(preparedGraph.nodes.map((node) => node.id));
    }

    return new Set(
      animationOrder.slice(0, Math.max(animationStep + 1, 0)).map((node) => node.id),
    );
  }, [animationOrder, animationStep, isAnimating, preparedGraph.nodes]);

  const visibleNodes = useMemo(() => {
    if (!isAnimating) {
      return preparedGraph.nodes;
    }

    return preparedGraph.nodes.filter((node) => visibleNodeIds.has(node.id));
  }, [isAnimating, preparedGraph.nodes, visibleNodeIds]);

  const visibleEdges = useMemo(() => {
    if (!isAnimating) {
      return preparedGraph.edges;
    }

    return preparedGraph.edges.filter(
      (edge) => visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target),
    );
  }, [isAnimating, preparedGraph.edges, visibleNodeIds]);

  useEffect(() => {
    setNodes(visibleNodes);
    setEdges(visibleEdges);
  }, [setEdges, setNodes, visibleEdges, visibleNodes]);

  useEffect(() => {
    if (
      filters.selectedCategoryId &&
      !categoryOptions.some((option) => option.id === filters.selectedCategoryId)
    ) {
      setFilters((current) => ({
        ...current,
        selectedCategoryId: "",
      }));
    }
  }, [categoryOptions, filters.selectedCategoryId]);

  useEffect(() => {
    stopAnimation(timerRef, setIsAnimating);
    setAnimationStep(-1);
  }, [preparedGraph]);

  useEffect(
    () => () => {
      if (timerRef.current) {
        window.clearInterval(timerRef.current);
        timerRef.current = null;
      }
    },
    [],
  );

  function handleNodeClick(_, node) {
    onOpenGraphNode?.(node.data);
  }

  function handleEdgeClick(event, edge) {
    if (edge?.data?.isUserLink) {
      event.preventDefault();
      void onDeleteGraphLink?.(edge.source, edge.target);
    }
  }

  function handleConnect(connection) {
    const sourceId = connection?.source;
    const targetId = connection?.target;
    if (!sourceId || !targetId || sourceId === targetId) {
      return;
    }

    void onCreateGraphLink?.(sourceId, targetId);
  }

  function handleStartAnimation() {
    if (loading || isAnimating || animationOrder.length === 0) {
      return;
    }

    stopAnimation(timerRef, setIsAnimating);

    setIsAnimating(true);
    setAnimationStep(0);

    if (animationOrder.length === 1) {
      setIsAnimating(false);
      return;
    }

    timerRef.current = window.setInterval(() => {
      setAnimationStep((currentStep) => {
        const nextStep = currentStep + 1;
        if (nextStep >= animationOrder.length - 1) {
          stopAnimation(timerRef, setIsAnimating);
          return animationOrder.length - 1;
        }

        return nextStep;
      });
    }, NODE_APPEAR_DELAY_MS);
  }

  function handleToggleFilter(key) {
    setFilters((current) => ({
      ...current,
      [key]: !current[key],
    }));
  }

  function handleSelectedCategoryChange(event) {
    const selectedCategoryId = event.target.value;
    setFilters((current) => ({
      ...current,
      selectedCategoryId,
    }));
  }

  function handleResetFilters() {
    setFilters({ ...DEFAULT_FILTERS });
  }

  return (
    <main className="graph-page">
      <header className="profile-header page-header">
        <div className="page-header-copy">
          <div className="page-header-leading">
            <WorkspaceContextMenu
              currentWorkspace={currentWorkspace}
              workspaces={workspaces}
              onSelectPersonalWorkspace={onSelectPersonalWorkspace}
              onSelectWorkspace={onSelectWorkspace}
              onCreateWorkspace={onCreateWorkspace}
            />
            <span className="eyebrow">{isWorkspaceMode ? "Общий режим" : "Личный режим"}</span>
          </div>
          <h1>{pageTitle}</h1>
          <p>Связанная база знаний, категории и заметки в одном визуальном слое.</p>
        </div>
        <div className="profile-actions page-header-actions">
          <button className="secondary-button" type="button" onClick={onBackToNotes}>
            Заметки
          </button>
          <button className="secondary-button" type="button" onClick={onOpenCalendar}>
            Календарь
          </button>
          <button className="secondary-button" type="button" onClick={onOpenProfile}>
            {contextButtonLabel}
          </button>
        </div>
      </header>

      <section className="graph-filters">
        <label className="graph-filter-toggle">
          <input
            type="checkbox"
            checked={filters.showCategories}
            onChange={() => handleToggleFilter("showCategories")}
          />
          <span>Категории</span>
        </label>

        <label className="graph-filter-toggle">
          <input
            type="checkbox"
            checked={filters.showNotes}
            onChange={() => handleToggleFilter("showNotes")}
          />
          <span>Заметки</span>
        </label>

        <label className="graph-filter-field">
          <span>Ветка категории</span>
          <select value={filters.selectedCategoryId} onChange={handleSelectedCategoryChange}>
            <option value="">Все категории</option>
            {categoryOptions.map((option) => (
              <option key={option.id} value={option.id}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <button
          className="secondary-button"
          type="button"
          onClick={handleResetFilters}
          disabled={!hasActiveFilters}
        >
          Сбросить
        </button>

        <button
          className="primary-button graph-filter-action"
          type="button"
          onClick={handleStartAnimation}
          disabled={loading || isAnimating || animationOrder.length === 0}
        >
          {isAnimating ? "Анимация идет..." : "Запустить анимацию"}
        </button>
      </section>

      {isAnimating ? (
        <div className="graph-status">
          {`Показано ${Math.min(animationStep + 1, animationOrder.length)} из ${animationOrder.length} узлов`}
        </div>
      ) : null}

      <section className="graph-panel">
        {loading && <div className="graph-loading">Загрузка графа...</div>}
        {!loading && preparedGraph.nodes.length === 0 && (
          <div className="graph-empty">
            <h2>Граф пустой</h2>
            <p>
              {hasActiveFilters
                ? "Текущие фильтры ничего не показали. Попробуйте другую категорию или сбросьте фильтры."
                : "Создайте категории и заметки, после этого они появятся здесь как узлы графа."}
            </p>
          </div>
        )}

        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={handleNodeClick}
          onEdgeClick={handleEdgeClick}
          onConnect={handleConnect}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          nodesDraggable
          nodesConnectable
          elementsSelectable
        >
          <Background color="#b8c2d8" gap={24} size={1} />
          <MiniMap pannable zoomable nodeStrokeWidth={3} />
          <Controls />
        </ReactFlow>
      </section>
    </main>
  );
}

function GraphNode({ data }) {
  const color = data.color || (data.type === "category" ? "#8FA3FF" : "#5ab9a9");

  return (
    <div className={`graph-node graph-node-${data.type}`} style={{ "--node-color": color }}>
      <Handle type="target" position={Position.Left} />
      <div className="graph-node-kind">{data.type === "category" ? "Категория" : "Заметка"}</div>
      <strong>{data.label || "Без названия"}</strong>
      {data.createdAtLabel ? <span className="graph-node-meta">{data.createdAtLabel}</span> : null}
      <Handle type="source" position={Position.Right} />
    </div>
  );
}

function toFlowGraph(graph) {
  const graphNodes = Array.isArray(graph?.nodes) ? graph.nodes : [];
  const graphEdges = Array.isArray(graph?.edges) ? graph.edges : [];

  const categories = graphNodes.filter((node) => node.type === "category");
  const notes = graphNodes.filter((node) => node.type === "note");
  const categoryEdges = graphEdges.filter((edge) => edge.type === "CHILD");
  const noteEdges = graphEdges.filter((edge) => edge.type !== "CHILD");

  const categoryPositions = buildCategoryPositions(categories, categoryEdges);
  const flowCategoryNodes = categories.map((category) =>
    toFlowNode(
      category,
      categoryPositions.get(category.id) || {
        x: CATEGORY_X,
        y: 0,
      },
    ),
  );

  const notesByCategory = new Map();
  notes.forEach((note) => {
    const key = note.category_uuid || "without-category";
    const items = notesByCategory.get(key) || [];
    items.push(note);
    notesByCategory.set(key, items);
  });

  const flowNoteNodes = notes.map((note, index) => {
    const siblings = notesByCategory.get(note.category_uuid || "without-category") || [];
    const siblingIndex = Math.max(
      siblings.findIndex((item) => item.id === note.id),
      0,
    );
    const categoryPosition = categoryPositions.get(note.category_uuid);

    return toFlowNode(note, {
      x: categoryPosition
        ? categoryPosition.x + NOTE_GAP_X + Math.floor(siblingIndex / 4) * (NODE_WIDTH + 80)
        : CATEGORY_X + NOTE_GAP_X,
      y: categoryPosition ? categoryPosition.y + siblingIndex * 88 : index * 100,
    });
  });

  const knownNodeIds = new Set([...flowCategoryNodes, ...flowNoteNodes].map((node) => node.id));
  const flowEdges = [...categoryEdges, ...noteEdges]
    .filter((edge) => knownNodeIds.has(edge.source) && knownNodeIds.has(edge.target))
    .map((edge, index) => ({
      id: `${edge.type}-${edge.source}-${edge.target}-${index}`,
      source: edge.source,
      target: edge.target,
      type: "smoothstep",
      markerEnd: {
        type: MarkerType.ArrowClosed,
      },
      className: resolveEdgeClassName(edge.type),
      data: {
        isUserLink: edge.type === "USER_LINK",
      },
      style: resolveEdgeStyle(edge.type),
    }));

  return {
    nodes: [...flowCategoryNodes, ...flowNoteNodes],
    edges: flowEdges,
  };
}

function buildCategoryPositions(categories, categoryEdges) {
  const categoryById = new Map(categories.map((category) => [category.id, category]));
  const childrenByParent = new Map();
  const incomingCount = new Map(categories.map((category) => [category.id, 0]));

  categoryEdges.forEach((edge) => {
    const children = childrenByParent.get(edge.source) || [];
    children.push(edge.target);
    childrenByParent.set(edge.source, children);
    incomingCount.set(edge.target, (incomingCount.get(edge.target) || 0) + 1);
  });

  const roots = categories
    .filter((category) => (incomingCount.get(category.id) || 0) === 0)
    .sort((left, right) => String(left.label || "").localeCompare(String(right.label || ""), "ru"));

  const positions = new Map();
  let rowIndex = 0;

  const visit = (categoryId, depth) => {
    if (positions.has(categoryId)) {
      return;
    }

    positions.set(categoryId, {
      x: CATEGORY_X + depth * CATEGORY_DEPTH_GAP,
      y: rowIndex * ROW_GAP,
    });
    rowIndex += 1;

    const childIds = (childrenByParent.get(categoryId) || [])
      .filter((childId) => categoryById.has(childId))
      .sort((left, right) =>
        String(categoryById.get(left)?.label || "").localeCompare(
          String(categoryById.get(right)?.label || ""),
          "ru",
        ),
      );

    childIds.forEach((childId) => visit(childId, depth + 1));
  };

  roots.forEach((category) => visit(category.id, 0));
  categories.forEach((category) => visit(category.id, 0));

  return positions;
}

function toFlowNode(node, position) {
  const createdAt = normalizeUnixTimestamp(node.created_at);

  return {
    id: node.id,
    type: "graphNode",
    position,
    data: {
      nodeId: node.id,
      type: node.type,
      label: node.label,
      color: node.color,
      category_uuid: node.category_uuid,
      createdAt,
      createdAtLabel: createdAt ? formatCreatedAt(createdAt) : "",
    },
  };
}

function compareAnimationNodes(left, right) {
  const leftTime = normalizeAnimationTime(left.data.createdAt);
  const rightTime = normalizeAnimationTime(right.data.createdAt);

  if (leftTime !== rightTime) {
    return leftTime - rightTime;
  }

  const leftTypePriority = left.data.type === "category" ? 0 : 1;
  const rightTypePriority = right.data.type === "category" ? 0 : 1;
  if (leftTypePriority !== rightTypePriority) {
    return leftTypePriority - rightTypePriority;
  }

  return String(left.id).localeCompare(String(right.id), "ru");
}

function normalizeAnimationTime(value) {
  if (!Number.isFinite(value) || value <= 0) {
    return Number.MAX_SAFE_INTEGER;
  }

  return value;
}

function resolveEdgeClassName(type) {
  if (type === "USER_LINK") {
    return "graph-edge-user";
  }
  return "graph-edge-category";
}

function resolveEdgeStyle(type) {
  if (type === "USER_LINK") {
    return {
      strokeWidth: 2.4,
    };
  }
  return {
    strokeWidth: 2.4,
  };
}

function normalizeUnixTimestamp(value) {
  if (!Number.isFinite(value) || value <= 0) {
    return null;
  }

  return Math.trunc(value);
}

function formatCreatedAt(value) {
  return new Date(value * 1000).toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function stopAnimation(timerRef, setIsAnimating) {
  if (timerRef.current) {
    window.clearInterval(timerRef.current);
    timerRef.current = null;
  }

  setIsAnimating(false);
}
