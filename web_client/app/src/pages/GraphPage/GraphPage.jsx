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
const NOTE_X = 390;
const ROW_GAP = 130;

const nodeTypes = {
  graphNode: GraphNode,
};

export default function GraphPage({
  graph,
  loading,
  onRefresh,
  onBackToNotes,
  onOpenProfile,
  onLogout,
}) {
  const preparedGraph = useMemo(() => toFlowGraph(graph), [graph]);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  useEffect(() => {
    setNodes(preparedGraph.nodes);
    setEdges(preparedGraph.edges);
  }, [preparedGraph, setEdges, setNodes]);

  return (
    <main className="graph-page">
      <header className="graph-header">
        <div>
          <span className="eyebrow">Граф заметок</span>
          <h1>Категории и заметки</h1>
          <p>Первый слой визуализации: категории, заметки и связи между ними. Узлы можно перетаскивать мышкой.</p>
        </div>
        <div className="graph-actions">
          <button className="secondary-button" type="button" onClick={onBackToNotes}>
            К заметкам
          </button>
          <button className="secondary-button" type="button" onClick={onOpenProfile}>
            Личный кабинет
          </button>
          <button className="secondary-button" type="button" onClick={onRefresh} disabled={loading}>
            Обновить
          </button>
          <button className="secondary-button" type="button" onClick={onLogout}>
            Выйти
          </button>
        </div>
      </header>

      <section className="graph-panel">
        {loading && <div className="graph-loading">Загрузка графа...</div>}
        {!loading && nodes.length === 0 && (
          <div className="graph-empty">
            <h2>Граф пока пустой</h2>
            <p>Создай категории и заметки, после этого они появятся здесь как узлы графа.</p>
          </div>
        )}

        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          nodesDraggable
          nodesConnectable={false}
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
  const categoryPositions = new Map();

  const flowCategoryNodes = categories.map((category, index) => {
    const position = {
      x: CATEGORY_X,
      y: index * ROW_GAP,
    };
    categoryPositions.set(category.id, position);

    return toFlowNode(category, position);
  });

  const notesByCategory = new Map();
  notes.forEach((note) => {
    const key = note.category_uuid || "without-category";
    const items = notesByCategory.get(key) || [];
    items.push(note);
    notesByCategory.set(key, items);
  });

  const flowNoteNodes = notes.map((note, index) => {
    const siblings = notesByCategory.get(note.category_uuid || "without-category") || [];
    const siblingIndex = Math.max(siblings.findIndex((item) => item.id === note.id), 0);
    const categoryPosition = categoryPositions.get(note.category_uuid);

    return toFlowNode(note, {
      x: NOTE_X + Math.floor(siblingIndex / 4) * (NODE_WIDTH + 80),
      y: categoryPosition ? categoryPosition.y + siblingIndex * 88 : index * 100,
    });
  });

  const knownNodeIds = new Set([...flowCategoryNodes, ...flowNoteNodes].map((node) => node.id));
  const flowEdges = graphEdges
    .filter((edge) => knownNodeIds.has(edge.source) && knownNodeIds.has(edge.target))
    .map((edge, index) => ({
      id: `${edge.type}-${edge.source}-${edge.target}-${index}`,
      source: edge.source,
      target: edge.target,
      label: edge.type === "HAS_NOTE" ? "содержит" : "связана",
      type: "smoothstep",
      markerEnd: {
        type: MarkerType.ArrowClosed,
      },
      className: edge.type === "LINKED_TO" ? "graph-edge-link" : "graph-edge-category",
    }));

  return {
    nodes: [...flowCategoryNodes, ...flowNoteNodes],
    edges: flowEdges,
  };
}

function toFlowNode(node, position) {
  return {
    id: node.id,
    type: "graphNode",
    position,
    data: {
      type: node.type,
      label: node.label,
      color: node.color,
      category_uuid: node.category_uuid,
    },
  };
}
