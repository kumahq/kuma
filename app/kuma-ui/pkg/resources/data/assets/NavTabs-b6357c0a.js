import{d as m,D as l,N as n,a as i,o as p,b as N,$ as v,J as f,w as u,e as b,f as h,t as x,q as y,aj as T,_ as k}from"./index-d015481a.js";const g=m({__name:"NavTabs",props:{tabs:{type:Array,required:!0}},setup(c){const o=l(),a=c,_=n(()=>a.tabs.map(t=>({title:t.title,hash:"#"+t.routeName}))),d=n(()=>{const t=o.matched.map(e=>e.meta.module??"").filter(e=>e!=="");t.reverse();const s=a.tabs.find(e=>!!(e.routeName===o.name||t.includes(e.module)));return"#"+((s==null?void 0:s.routeName)??a.tabs[0].routeName)});return(t,s)=>{const r=i("RouterLink");return p(),N(y(T),{tabs:_.value,"model-value":d.value,"has-panels":!1,class:"nav-tabs","data-testid":"nav-tabs"},v({_:2},[f(a.tabs,e=>({name:`${e.routeName}-anchor`,fn:u(()=>[b(r,{"data-testid":`${e.routeName}-tab`,to:{name:e.routeName}},{default:u(()=>[h(x(e.title),1)]),_:2},1032,["data-testid","to"])])}))]),1032,["tabs","model-value"])}}});const L=k(g,[["__scopeId","data-v-d59352a2"]]);export{L as N};
