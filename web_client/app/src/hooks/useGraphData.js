import { useEffect, useState } from "react";

import { fetchGraph } from "../api/graphApi";

const EMPTY_GRAPH = {
  nodes: [],
  edges: [],
};

export function useGraphData({ token, enabled, setMessage }) {
  const [graph, setGraph] = useState(EMPTY_GRAPH);
  const [graphLoading, setGraphLoading] = useState(false);

  useEffect(() => {
    setGraph(EMPTY_GRAPH);
  }, [token]);

  useEffect(() => {
    if (!enabled || !token) {
      return;
    }

    void loadGraph();
  }, [enabled, token]);

  async function loadGraph() {
    if (!token) {
      return;
    }

    try {
      setGraphLoading(true);
      const nextGraph = await fetchGraph(token);
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

  return {
    graph,
    graphLoading,
    refreshGraph: loadGraph,
  };
}
