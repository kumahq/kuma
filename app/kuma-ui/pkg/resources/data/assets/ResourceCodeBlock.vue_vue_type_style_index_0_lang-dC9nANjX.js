import{C as T}from"./CodeBlock-3HK16lYA.js";import{d as _,l as k,M as B,C as K,a as g,o as w,c as q,e as l,w as i,p,ae as E,t as R,f as x,r as S}from"./index-i9Uphcre.js";import{t as F}from"./toYaml-sPaYOD3i.js";const V=_({__name:"ResourceCodeBlock",props:{resource:{},codeMaxHeight:{default:void 0},isSearchable:{type:Boolean,default:!1},query:{default:""},isFilterMode:{type:Boolean,default:!1},isRegExpMode:{type:Boolean,default:!1}},emits:["query-change","filter-mode-change","reg-exp-mode-change"],setup(h,{emit:b}){const{t:c}=k(),t=h,u=b,C=B(()=>f(t.resource)),m=K(()=>{});let d=new Promise((a,e)=>{m.value=n=>n(a,e)});const v=async a=>{let e;try{e=await d}finally{d=new Promise((n,s)=>{m.value=r=>r(n,s)})}return e};async function M(){return f(await v())}function f(a){const{creationTime:e,modificationTime:n,...s}=a;return F(s)}return(a,e)=>{const n=g("KTooltip"),s=g("KToggle");return w(),q("div",null,[l(s,{toggled:!1},{default:i(({isToggled:r,toggle:y})=>[l(T,{language:"yaml",code:C.value,"is-searchable":t.isSearchable,"code-max-height":t.codeMaxHeight,query:t.query,"is-filter-mode":t.isFilterMode,"is-reg-exp-mode":t.isRegExpMode,onQueryChange:e[0]||(e[0]=o=>u("query-change",o)),onFilterModeChange:e[1]||(e[1]=o=>u("filter-mode-change",o)),onRegExpModeChange:e[2]||(e[2]=o=>u("reg-exp-mode-change",o))},{"secondary-actions":i(()=>[l(n,{class:"kubernetes-copy-button-tooltip",text:p(c)("common.copyKubernetesText"),placement:"bottomEnd","max-width":"200"},{default:i(()=>[l(E,{class:"kubernetes-copy-button","get-text":M,"copy-text":p(c)("common.copyKubernetesText"),"has-border":"","hide-title":"","icon-color":"currentColor",onClick:()=>{r.value===!1&&y()}},{default:i(()=>[x(R(p(c)("common.copyKubernetesShortText")),1)]),_:2},1032,["copy-text","onClick"])]),_:2},1032,["text"])]),_:2},1032,["code","is-searchable","code-max-height","query","is-filter-mode","is-reg-exp-mode"]),x(),S(a.$slots,"default",{copy:o=>{r.value!==!1&&y(),m.value(o)},copying:r.value})]),_:3})])}}});export{V as _};
