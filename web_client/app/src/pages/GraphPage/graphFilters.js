const DEFAULT_GRAPH = {
  nodes: [],
  edges: [],
};

export function filterGraph(graph, filters) {
  const graphNodes = Array.isArray(graph?.nodes) ? graph.nodes : DEFAULT_GRAPH.nodes;
  const graphEdges = Array.isArray(graph?.edges) ? graph.edges : DEFAULT_GRAPH.edges;
  const nodesById = new Map(graphNodes.map((node) => [node.id, node]));

  let branchVisibleNodeIds = null;
  if (filters.selectedCategoryId) {
    const branchCategoryIds = collectCategoryBranch(
      filters.selectedCategoryId,
      graphNodes,
      graphEdges,
    );

    if (branchCategoryIds) {
      branchVisibleNodeIds = collectVisibleNodeIdsForBranch(
        branchCategoryIds,
        graphNodes,
        graphEdges,
      );
      includeParentCategoriesForVisibleNotes(branchVisibleNodeIds, nodesById);
    }
  }

  const filteredNodes = graphNodes.filter((node) => {
    if (branchVisibleNodeIds && !branchVisibleNodeIds.has(node.id)) {
        return false;
    }

    if (node.type === "category" && !filters.showCategories) {
      return false;
    }

    if (node.type === "note" && !filters.showNotes) {
      return false;
    }

    return true;
  });

  const visibleNodeIds = new Set(filteredNodes.map((node) => node.id));
  const filteredEdges = graphEdges.filter(
    (edge) => visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target),
  );

  return {
    nodes: filteredNodes,
    edges: filteredEdges,
  };
}

export function buildCategoryOptions(graph) {
  const graphNodes = Array.isArray(graph?.nodes) ? graph.nodes : DEFAULT_GRAPH.nodes;
  const graphEdges = Array.isArray(graph?.edges) ? graph.edges : DEFAULT_GRAPH.edges;
  const categories = graphNodes.filter((node) => node.type === "category");
  const categoryById = new Map(categories.map((category) => [category.id, category]));
  const childrenByParent = buildChildrenMap(graphEdges, categoryById);
  const incomingCount = new Map(categories.map((category) => [category.id, 0]));

  graphEdges
    .filter((edge) => edge.type === "CHILD")
    .forEach((edge) => incomingCount.set(edge.target, (incomingCount.get(edge.target) || 0) + 1));

  const roots = categories
    .filter((category) => (incomingCount.get(category.id) || 0) === 0)
    .sort(compareCategoryLabels);

  const options = [];
  const visited = new Set();

  const visit = (categoryId, depth) => {
    if (visited.has(categoryId) || !categoryById.has(categoryId)) {
      return;
    }

    visited.add(categoryId);

    const category = categoryById.get(categoryId);
    options.push({
      id: category.id,
      label: `${"> ".repeat(depth)}${category.label || "Без названия"}`,
    });

    const children = [...(childrenByParent.get(categoryId) || [])].sort((left, right) =>
      compareCategoryLabels(categoryById.get(left), categoryById.get(right)),
    );
    children.forEach((childId) => visit(childId, depth + 1));
  };

  roots.forEach((category) => visit(category.id, 0));
  categories.forEach((category) => visit(category.id, 0));

  return options;
}

function collectVisibleNodeIdsForBranch(branchCategoryIds, graphNodes, graphEdges) {
  const branchNodeIds = new Set();

  graphNodes.forEach((node) => {
    if (node.type === "category" && branchCategoryIds.has(node.id)) {
      branchNodeIds.add(node.id);
      return;
    }

    if (node.type === "note" && branchCategoryIds.has(node.category_uuid)) {
      branchNodeIds.add(node.id);
    }
  });

  return includeDirectUserLinkNeighbours(branchNodeIds, graphEdges);
}

function collectCategoryBranch(rootId, graphNodes, graphEdges) {
  const categories = graphNodes.filter((node) => node.type === "category");
  const categoryById = new Map(categories.map((category) => [category.id, category]));
  if (!categoryById.has(rootId)) {
    return null;
  }

  const childrenByParent = buildChildrenMap(graphEdges, categoryById);
  const branchIds = new Set();
  const stack = [rootId];

  while (stack.length > 0) {
    const categoryId = stack.pop();
    if (!categoryId || branchIds.has(categoryId)) {
      continue;
    }

    branchIds.add(categoryId);
    const children = childrenByParent.get(categoryId) || [];
    children.forEach((childId) => {
      if (!branchIds.has(childId)) {
        stack.push(childId);
      }
    });
  }

  return branchIds;
}

function includeDirectUserLinkNeighbours(branchNodeIds, graphEdges) {
  const visibleNodeIds = new Set(branchNodeIds);
  const neighboursByNodeId = buildUserLinkMap(graphEdges);

  branchNodeIds.forEach((nodeId) => {
    const neighbours = neighboursByNodeId.get(nodeId) || [];
    neighbours.forEach((neighbourId) => visibleNodeIds.add(neighbourId));
  });

  return visibleNodeIds;
}

function includeParentCategoriesForVisibleNotes(visibleNodeIds, nodesById) {
  const visibleIdsSnapshot = [...visibleNodeIds];

  visibleIdsSnapshot.forEach((nodeId) => {
    const node = nodesById.get(nodeId);
    if (node?.type !== "note" || !node.category_uuid) {
      return;
    }

    if (nodesById.has(node.category_uuid)) {
      visibleNodeIds.add(node.category_uuid);
    }
  });
}

function buildChildrenMap(graphEdges, categoryById) {
  const childrenByParent = new Map();

  graphEdges
    .filter((edge) => edge.type === "CHILD")
    .forEach((edge) => {
      if (!categoryById.has(edge.source) || !categoryById.has(edge.target)) {
        return;
      }

      const children = childrenByParent.get(edge.source) || [];
      children.push(edge.target);
      childrenByParent.set(edge.source, children);
    });

  return childrenByParent;
}

function buildUserLinkMap(graphEdges) {
  const neighboursByNodeId = new Map();

  graphEdges
    .filter((edge) => edge.type === "USER_LINK")
    .forEach((edge) => {
      appendNeighbour(neighboursByNodeId, edge.source, edge.target);
      appendNeighbour(neighboursByNodeId, edge.target, edge.source);
    });

  return neighboursByNodeId;
}

function appendNeighbour(neighboursByNodeId, sourceId, targetId) {
  if (!sourceId || !targetId) {
    return;
  }

  const neighbours = neighboursByNodeId.get(sourceId) || [];
  neighbours.push(targetId);
  neighboursByNodeId.set(sourceId, neighbours);
}

function compareCategoryLabels(left, right) {
  return String(left?.label || "").localeCompare(String(right?.label || ""), "ru");
}
