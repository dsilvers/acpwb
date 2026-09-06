package archive

import "testing"

func BenchmarkRenderArchiveDefault(b *testing.B) {
	ctx := BuildContext(2010, 6, 15, "bench-archive-slug")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderArchiveDefault(&ctx)
	}
}
