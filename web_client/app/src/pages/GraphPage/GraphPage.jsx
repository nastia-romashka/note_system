import { useEffect, useMemo } from "react";
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

const NODE_WIDTH = 220;
const CATEGORY_X = 40;
const CATEGORY_DEPTH_GAP = 280;
const NOTE_GAP_X = 350;
const ROW_GAP = 130;

const nodeTypes = {
  graphNode: GraphNode,
};

export default function GraphPage({
  graph,
  loading,
  onCreateGraphLink,
  onDeleteGraphLink,
  onBackToNotes,
  onOpenCalendar,
  onOpenGraphNode,
  onOpenProfile,
}) {
  const preparedGraph = useMemo(() => toFlowGraph(graph), [graph]);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  useEffect(() => {
    setNodes(preparedGraph.nodes);
    setEdges(preparedGraph.edges);
  }, [preparedGraph, setEdges, setNodes]);

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

  return (
    <main className="graph-page">
      <header className="graph-header">
        <div>
          <span className="eyebrow">Связанная база знаний</span>
          <h1>Граф</h1>
        </div>
        <div className="graph-actions">
          <button className="secondary-button" type="button" onClick={onBackToNotes}>
            К заметкам
          </button>
          <button className="secondary-button" type="button" onClick={onOpenCalendar}>
            Календарь
          </button>
          <button className="secondary-button" type="button" onClick={onOpenProfile}>
            Личный кабинет
          </button>
        </div>
      </header>

      <section className="graph-panel">
        {loading && <div className="graph-loading">Загрузка графа...</div>}
        {!loading && nodes.length === 0 && (
          <div className="graph-empty">
            <h2>Граф пока пустой</h2>
            <p>Создайте категории и заметки, после этого они появятся здесь как узлы графа.</p>
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
    },
  };
}

function resolveEdgeClassName(type) {
  if (type === "USER_LINK") {
    return "graph-edge-user";
  }
  if (type === "LINKED_TO") {
    return "graph-edge-link";
  }
  return "graph-edge-category";
}

function resolveEdgeStyle(type) {
  if (type === "USER_LINK") {
    return {
      strokeWidth: 2.4,
    };
  }
  if (type === "LINKED_TO") {
    return {
      strokeWidth: 2.2,
    };
  }
  return {
    strokeWidth: 2.4,
  };
}
