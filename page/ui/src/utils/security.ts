// 将密级脱敏字符转为文字描述。
// all-**** ：替换为****；each-***: 每个字符替换为***；each-^**：每2个字符替换为1个*；start-***：前3个字符替换为***；end-***：后3个字符替换为***；middle-***: 中间3个字符替换为***；
// -：不替换
// 其中的*是可变的，数量表示替换数量
export function securityLevelToText(security?: string): string {
    if (!security) return "未配置";
    if (security === "-") return "不脱敏";
    const ss = security.split("-");
    if (ss.length === 1) return "未配置";
    const t = ss[0];
    // 是否有^
    const t1 = ss[1];
    const hasArrow = t1.indexOf("^") > -1;
    // 星号数量
    let starCount = t1.length;
    if (hasArrow) {
        starCount--;
    }
    switch (t) {
        case "all":
            return "替换为" + (hasArrow ? t1.substring(1) : t1);
        case "each":
            if (hasArrow) {
                return `每${starCount}个字符替换为1个*`;
            } else {
                return `每个字符替换为${starCount}个*`;
            }
        case "start":
            return `前${starCount}个字符替换为*`;
        case "end":
            return `后${starCount}个字符替换为*`;
        case "middle":
            return `中间${starCount}个字符替换为*`;
        default:
            return "未配置";
    }
}
