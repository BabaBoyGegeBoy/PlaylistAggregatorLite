package logic

import "testing"

// sig 取两首歌的签名并判断是否应合并（同核心歌名 + 歌手兼容）
func shouldMerge(t *testing.T, a, b string) bool {
	t.Helper()
	sa, sb := songSignature(a, ""), songSignature(b, "")
	return sa.core == sb.core && artistCompatible(sa, sb)
}

func TestDedupSignature(t *testing.T) {
	// 期望合并（核心歌名 + 歌手集 兼容）
	merge := [][2]string{
		// 原有样本
		{"真夜中のドア/Stay With Me (深夜门扉/留在我身边)(シングルver.) - 松原みき", "真夜中のドア/Stay With Me (シングルver.) - 松原みき"},
		{"blue (with MINNIE) - yung kai", "blue - yung kai&MINNIE"},
		{"blue (with MINNIE) - yung kai, MINNIE", "blue (with MINNIE) - yung kai&MINNIE"},
		{"The Other Side Of Paradise (Explicit) - X", "The Other Side Of Paradise - X"},
		{"Tek It (Acoustic) - X", "Tek It - X"},                 // acoustic 属轻度信息，去除后合并
		{"하늘을 날다 - 에일리 (AILEE)", "하늘을 날다 - AILEE"}, // 艺人括号别名替换基名
		{"ＴＲＡ＄Ｈ - Z", "TRA$H - Z"},                         // 全角转半角

		// 用户 9 组需求
		{"真夜中のドア/Stay With Me (シングルver.) - 松原みき", "真夜中のドア〜stay with me (シングルver.) - 松原みき"},                 // #1 波浪号/分隔符归一
		{"Are You Lost - Park Bird", "Are You Lost (你迷失了吗) - 未知艺人"},                                       // #2 占位歌手通配
		{"Born a Stranger (生而陌路) - Kan Gao", "Born a Stranger (生而陌路) - Kan Gao,Laura Shigihara"},         // #3 歌手集子集
		{"Don't Look Back (feat. Kotomi & Ryan Elder) [From Rick and Morty: Season 4] - RICK AND MORTY,Kotomi,Ryan Elder",
			"Don't Look Back (feat. Kotomi & Ryan Elder) [From Rick and Morty: Season 4] - 瑞克和莫蒂,Kotomi,Ryan Elder"}, // #4 艺人别名
		{"I Don't Want to Play Around - Ace Spectrum", "I Don't Want to Play Around - Ace Specturm"},             // #5 拼写错别名
		{"In Love - 줄라이", "In Love - July"},                                                                    // #6 韩文->英文别名
		{"Luv(sic.) (Instrumental) - Nujabes", "Luv (sic)(Instrumentals) - Nujabes"},                             // #7 三选一(1-2)
		{"Luv (sic)(Instrumentals) - Nujabes", "Luv (sic. Instrumental) - Nujabes"},                              // #7 三选一(2-3)
		{"Promise - 山岡晃", "Promise - 山冈晃"},                                                                  // #8 日文汉字->简体
		{"Somebody That I Used To Know - Gotye,Kimbra", "Somebody That I Used to Know - Gotye"},                  // #9 歌手集子集
	}
	for _, p := range merge {
		if !shouldMerge(t, p[0], p[1]) {
			t.Errorf("应合并却未合并:\n  %q\n  %q", p[0], p[1])
		}
	}

	// 期望不合并（签名不同）
	diff := [][2]string{
		{"Live Forever - Oasis", "Live Forever (feat. Kotomi & Ryan Elder)[from Rick and Morty: Season 4] - Rick and Morty"},
		{"Try - Colbie Caillat", "Try - P!NK"},
		{"Luv (sic) - Nujabes", "Luv (sic) - Shing02"},
		{"Yesterday Once More (1991 Remix) - X", "Yesterday Once More - X"}, // remix 保留
		{"Wasted (Nightcore) - A", "Wasted - A"},                            // nightcore 保留
		{"Luv (sic. Instrumental) - X", "Luv (sic) - X"},                    // instrumental 保留(无 instrumental 者不同)
	}
	for _, p := range diff {
		if shouldMerge(t, p[0], p[1]) {
			t.Errorf("不应合并却合并了:\n  %q\n  %q", p[0], p[1])
		}
	}
}

// TestDedupKeep 验证合并时保留的代表条符合用户偏好（优先级更高）
func TestDedupKeep(t *testing.T) {
	keep := [][2]string{
		{"Are You Lost - Park Bird", "Are You Lost (你迷失了吗) - 未知艺人"},                                       // #2 留真实歌手
		{"Born a Stranger (生而陌路) - Kan Gao,Laura Shigihara", "Born a Stranger (生而陌路) - Kan Gao"},         // #3 留全集
		{"Don't Look Back (feat. Kotomi & Ryan Elder) [From Rick and Morty: Season 4] - RICK AND MORTY,Kotomi,Ryan Elder",
			"Don't Look Back (feat. Kotomi & Ryan Elder) [From Rick and Morty: Season 4] - 瑞克和莫蒂,Kotomi,Ryan Elder"}, // #4 留规范英文
		{"I Don't Want to Play Around - Ace Spectrum", "I Don't Want to Play Around - Ace Specturm"},             // #5 留规范拼写
		{"In Love - July", "In Love - 줄라이"},                                                                    // #6 留英文
		{"Luv(sic.) (Instrumental) - Nujabes", "Luv (sic)(Instrumentals) - Nujabes"},                             // #7 留 (sic.) 规范写法
		{"Promise - 山冈晃", "Promise - 山岡晃"},                                                                  // #8 留简体
		{"Somebody That I Used To Know - Gotye,Kimbra", "Somebody That I Used to Know - Gotye"},                  // #9 留全集
	}
	for _, p := range keep {
		want, other := songSignature(p[0], ""), songSignature(p[1], "")
		if want.priority <= other.priority {
			t.Errorf("保留优先级不符合预期:\n  期望优先: %q (prio=%d)\n  实际更高: %q (prio=%d)",
				p[0], want.priority, p[1], other.priority)
		}
	}
}
