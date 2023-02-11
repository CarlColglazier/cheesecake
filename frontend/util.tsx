function quartile(points: any[], counts: number[], q: number) {
    const all_count = counts.reduce(function (a, b) { return a + b }, 0);
    let c = 0;
    for (let i = 0; i < points.length; i++) {
      c += counts[i];
      if (c > all_count*q) {
        return points[i];
      }
    }
  }

  export default quartile;