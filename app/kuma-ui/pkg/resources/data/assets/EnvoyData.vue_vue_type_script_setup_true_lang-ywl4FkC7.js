import{d as y,a as p,o as n,c as _,e as i,w as l,b as a,a3 as k,a1 as x,$ as C,p as B,a2 as h,f as q}from"./index-CImj3nNu.js";import{C as E}from"./CodeBlock-CsNe8TyX.js";const D=y({__name:"EnvoyData",props:{resource:{},src:{},query:{default:""},isFilterMode:{type:Boolean,default:!1},isRegExpMode:{type:Boolean,default:!1}},emits:["query-change","filter-mode-change","reg-exp-mode-change"],setup(m,{emit:d}){const s=m,t=d;return(M,e)=>{const u=p("KButton"),f=p("DataSource");return n(),_("div",null,[i(f,{src:s.src},{default:l(({data:o,error:c,refresh:g})=>[c?(n(),a(k,{key:0,error:c},null,8,["error"])):o===void 0?(n(),a(x,{key:1})):o===""?(n(),a(C,{key:2})):(n(),a(E,{key:3,language:"json",code:typeof o=="string"?o:JSON.stringify(o,null,2),"is-searchable":"",query:s.query,"is-filter-mode":s.isFilterMode,"is-reg-exp-mode":s.isRegExpMode,onQueryChange:e[0]||(e[0]=r=>t("query-change",r)),onFilterModeChange:e[1]||(e[1]=r=>t("filter-mode-change",r)),onRegExpModeChange:e[2]||(e[2]=r=>t("reg-exp-mode-change",r))},{"primary-actions":l(()=>[i(u,{appearance:"primary",onClick:g},{default:l(()=>[i(B(h)),q(`

            Refresh
          `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode"]))]),_:1},8,["src"])])}}});export{D as _};
