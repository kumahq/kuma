import{d as B,l as M,D as r,H as T,a as b,o as q,c as w,e as c,w as g,p as u,ae as E,af as R,f as x,t as S,r as K}from"./index-CImj3nNu.js";import{C as _}from"./CodeBlock-CsNe8TyX.js";import{t as D}from"./toYaml-DB9FPXFY.js";const N=B({__name:"ResourceCodeBlock",props:{resource:{},codeMaxHeight:{default:void 0},isSearchable:{type:Boolean,default:!1},query:{default:""},isFilterMode:{type:Boolean,default:!1},isRegExpMode:{type:Boolean,default:!1}},emits:["query-change","filter-mode-change","reg-exp-mode-change"],setup(h,{emit:v}){const{t:m}=M(),t=h,l=v,s=r(!1),d=r(""),p=r(null),C=T(()=>y(t.resource)),i=r(()=>{});let f=new Promise((a,o)=>{i.value=n=>n(a,o)});const k=async a=>{try{return f}finally{f=new Promise((o,n)=>{i.value=e=>e(o,n)})}};function y(a){const{creationTime:o,modificationTime:n,...e}=a;return D(e)}return(a,o)=>{const n=b("KCodeBlockIconButton");return q(),w("div",null,[c(_,{language:"yaml",code:C.value,"is-searchable":t.isSearchable,"code-max-height":t.codeMaxHeight,query:t.query,"is-filter-mode":t.isFilterMode,"is-reg-exp-mode":t.isRegExpMode,onQueryChange:o[1]||(o[1]=e=>l("query-change",e)),onFilterModeChange:o[2]||(o[2]=e=>l("filter-mode-change",e)),onRegExpModeChange:o[3]||(o[3]=e=>l("reg-exp-mode-change",e))},{"secondary-actions":g(()=>[c(n,{"copy-tooltip":u(m)("common.copyKubernetesText"),theme:"dark",onClick:o[0]||(o[0]=async()=>{var e;s.value||(s.value=!0,d.value=y(await k()),await E(),(e=p.value)==null||e.copy())})},{default:g(()=>[c(u(R),{ref_key:"kCopyElement",ref:p,format:"hidden",text:d.value},null,8,["text"]),x(" "+S(u(m)("common.copyKubernetesShortText")),1)]),_:1},8,["copy-tooltip"])]),_:1},8,["code","is-searchable","code-max-height","query","is-filter-mode","is-reg-exp-mode"]),x(),K(a.$slots,"default",{copy:e=>{s.value&&(s.value=!1),i.value(e)},copying:s.value})])}}});export{N as _};
