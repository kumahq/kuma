import{d as v,B as b,j as p,I as y,o as _,a as k,A as $,s as B,w,e as S,y as g,b as h,K as z,z as A,g as C,F as x,q as N}from"./index-065c0e80.js";import{f as q,J as E}from"./RouteView.vue_vue_type_script_setup_true_lang-1d679e8a.js";const K={key:0,class:"app-collection-toolbar"},L=v({__name:"AppCollection",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{}},emits:["change"],setup(d,{emit:i}){const o=d,u=b(),a=p(o.items),s=p(0);y(()=>o.items,()=>{s.value++,a.value=o.items});const r=n=>{const c=n.target.closest("tr");if(c){const t=c.querySelector("a");t!==null&&t.click()}};return(n,c)=>(_(),k(h(z),{class:"app-collection","has-error":typeof o.error<"u","pagination-total-items":o.total,"initial-fetcher-params":{page:o.pageNumber,pageSize:o.pageSize},"fetcher-cache-key":String(s.value),fetcher:({page:t,pageSize:l,query:e})=>(i("change",{page:t,size:l,s:e}),{data:a.value}),"cell-attrs":({headerKey:t})=>({class:`${t}-column`}),"empty-state-icon-size":"96","disable-sorting":"","hide-pagination-when-optional":"","onRow:click":r},$({_:2},[B(Object.keys(h(u)),t=>({name:t,fn:w(({row:l,rowValue:e})=>[t==="toolbar"?(_(),S("div",K,[g(n.$slots,"toolbar",{},void 0,!0)])):g(n.$slots,t,{key:1,row:l,rowValue:e},void 0,!0)])}))]),1032,["has-error","pagination-total-items","initial-fetcher-params","fetcher-cache-key","fetcher","cell-attrs"]))}});const I=q(L,[["__scopeId","data-v-f33076d8"]]),j=N("span",null,null,-1),V=v({__name:"DataSource",props:{src:{type:String,required:!0}},emits:["change","error"],setup(d,{emit:i}){const o=d,u=E(),a=p(void 0),s=p(void 0);let r={};const n=Symbol(""),c=async e=>{if(a.value=void 0,r=t(r),r.src=e,e==="")return;r.controller=new AbortController;const f=u.source(e,n);f.addEventListener("message",m=>{a.value=m.data,s.value=void 0,i("change",a.value)},{signal:r.controller.signal}),f.addEventListener("error",m=>{s.value=m.error,i("error",s.value)},{signal:r.controller.signal})},t=e=>(typeof e.controller<"u"&&e.controller.abort(),typeof e.src<"u"&&u.close(e.src,n),{});y(()=>o.src,e=>c(e),{immediate:!0}),A(()=>{r=t(r)});const l=()=>{c(o.src)};return(e,f)=>(_(),S(x,null,[g(e.$slots,"default",{data:a.value,error:s.value,refresh:l}),C(),j],64))}});export{I as A,V as _};
