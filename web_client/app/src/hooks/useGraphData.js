import { useEffect, useState } from "react";

import { createGraphLink, deleteGraphLink, fetchGraph } from "../api/graphApi";

const EMPTY_GRAPH = {
  nodes: [],
  edges: [],
};

export function useGraphData({ token, workspaceId, enabled, setMessage }) {
  const [graph, setGraph] = useState(EMPTY_GRAPH);
  const [graphLoading, setGraphLoading] = useState(false);

  useEffect(() => {
    setGraph(EMPTY_GRAPH);
  }, [token, workspaceId]);

  useEffect(() => {
    if (!enabled || !token) {
      return;
    }

    void loadGraph();
  }, [enabled, token, workspaceId]);

  async function loadGraph() {
    if (!token) {
      return;
    }

    try {
      setGraphLoading(true);
      const nextGraph = await fetchGraph(token, workspaceId);
      setGraph({
        nodes: Array.isArray(nextGraph?.nodes) ? nextGraph.nodes : [],
        edges: Array.isArray(nextGraph?.edges) ? nextGraph.edges : [],
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось загрузить граф.",
      });
    } finally {
      setGraphLoading(false);
    }
  }

  async function createUserGraphLink(sourceId, targetId) {
    if (!token || !sourceId || !targetId) {
      return;
    }

    try {
      setGraphLoading(true);
      await createGraphLink(token, {
        source_id: sourceId,
        target_id: targetId,
      }, workspaceId);
      await loadGraph();
      setMessage({
        type: "success",
        text: "Пользовательская связь добавлена.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось создать связь.",
      });
    } finally {
      setGraphLoading(false);
    }
  }

  async function deleteUserGraphLink(sourceId, targetId) {
    if (!token || !sourceId || !targetId) {
      return;
    }

    try {
      setGraphLoading(true);
      await deleteGraphLink(token, {
        source_id: sourceId,
        target_id: targetId,
      }, workspaceId);
      await loadGraph();
      setMessage({
        type: "success",
        text: "Пользовательская связь удалена.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось удалить связь.",
      });
    } finally {
      setGraphLoading(false);
    }
  }

  return {
    graph,
    graphLoading,
    createUserGraphLink,
    deleteUserGraphLink,
  };
}
