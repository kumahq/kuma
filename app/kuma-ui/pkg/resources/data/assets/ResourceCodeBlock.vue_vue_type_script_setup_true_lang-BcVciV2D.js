import{d as T,k as K,M as R,a2 as I,r as a,o as m,m as p,w as s,b as r,p as u,e as f,t as S,q as v,a as w}from"./index-DIMZSgEC.js";const E=T({__name:"ResourceCodeBlock",props:{resource:{},codeMaxHeight:{default:void 0},isSearchable:{type:Boolean,default:!1},query:{default:""},isFilterMode:{type:Boolean,default:!1},isRegExpMode:{type:Boolean,default:!1},showK8sCopyButton:{type:Boolean,default:!0}},emits:["query-change","filter-mode-change","reg-exp-mode-change"],setup(y,{emit:g}){const{t:c}=K(),n=y,i=g,C=R(()=>l(n.resource));function l(e){return"creationTime"in e&&delete e.creationTime,"modificationTime"in e&&delete e.modificationTime,I.stringify(e)}const h=e=>console.error(e);return(e,t)=>{const B=a("XIcon"),x=a("KCodeBlockIconButton"),k=a("XCopyButton"),_=a("XDisclosure"),M=a("XCodeBlock");return m(),p(M,{language:"yaml",code:C.value,"is-searchable":n.isSearchable,"code-max-height":n.codeMaxHeight,query:n.query,"is-filter-mode":n.isFilterMode,"is-reg-exp-mode":n.isRegExpMode,onQueryChange:t[0]||(t[0]=o=>i("query-change",o)),onFilterModeChange:t[1]||(t[1]=o=>i("filter-mode-change",o)),onRegExpModeChange:t[2]||(t[2]=o=>i("reg-exp-mode-change",o))},{"secondary-actions":s(()=>[r(_,null,{default:s(({expanded:o,toggle:d})=>[n.showK8sCopyButton?(m(),p(x,{key:0,"copy-tooltip":u(c)("common.copyKubernetesText"),theme:"dark",onClick:()=>{o||d()}},{default:s(()=>[r(B,{name:"copy"}),f(S(u(c)("common.copyKubernetesShortText")),1)]),_:2},1032,["copy-tooltip","onClick"])):v("",!0),t[3]||(t[3]=f()),r(k,{format:"hidden"},{default:s(({copy:X})=>[w(e.$slots,"default",{copy:b=>{o&&d(),b(q=>X(l(q)),h)},copying:o})]),_:2},1024)]),_:3})]),_:3},8,["code","is-searchable","code-max-height","query","is-filter-mode","is-reg-exp-mode"])}}});export{E as _};
