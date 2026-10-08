import assert from "node:assert/strict";

// pipe() unpipes and pauses its source when the destination closes. Keep the
// independent data/end observer flowing after a browser abandons its response.
// Install AFTER pipe's close listener so its unpipe cannot undo resume().
export function pipeObservedAPIResponse(received, response) {
  if (response.destroyed) { received.resume(); return; }
  received.pipe(response);
  response.once("close", () => { received.resume(); });
}

// A page reload may abandon old-page reads. This is not consumed-response
// evidence: the proxy must still enumerate their complete upstream bytes.
export function navigationAPIProof() {
  const boundaries = [], navigations = [];
  let navigation = null, sequence = 0;
  return {
    boundaries, navigations,
    observe(response, path, method, documentURL) {
      const row = { path, method, status: null, upstream_ended: false,
        browser_finished: false, browser_closed: false, cancelled_by_navigation: null,
        document_url: documentURL, started_ms: Date.now(), closed_ms: null,
        started_during_navigation: navigation?.id ?? null };
      boundaries.push(row);
      response.on("finish", () => { row.browser_finished = true; });
      response.on("close", () => {
        row.browser_closed = true;
        row.closed_ms = Date.now();
        if (!row.browser_finished && navigation && row.document_url === navigation.oldDocument &&
            row.method === "GET" && ["/api/v1/founder/state", "/api/v1/garden/current"].includes(row.path)) {
          row.cancelled_by_navigation = navigation.id;
        }
      });
      return row;
    },
    async reload(oldDocument, action) {
      assert.equal(navigation, null, "nested observer navigation");
      assert.equal(typeof oldDocument, "string", "navigation needs the actual old document URL");
      navigation = { id: ++sequence, oldDocument };
      const marker = { id: navigation.id, old_document: oldDocument, started_ms: Date.now(), finished_ms: null };
      navigations.push(marker);
      try { return await action(); }
      finally { marker.finished_ms = Date.now(); navigation = null; }
    },
    finish() {
      const cancelled = (row) => row.cancelled_by_navigation !== null && row.browser_closed && !row.browser_finished;
      const incomplete = boundaries.filter((row) => row.status === null || !row.upstream_ended ||
        !row.browser_finished && !cancelled(row));
      assert.equal(incomplete.length, 0, `API proxy observation incomplete: ${JSON.stringify({ incomplete, navigations })}`);
      return { upstream_responses: boundaries.length, navigation_cancelled_reads: boundaries.filter(cancelled).length };
    },
  };
}
