import{d as l,z as _,G as n,a as i,o as p,b as N,V as f,C as v,w as u,e as b,f as h,t as k,l as x,aj as y,_ as T}from"./index-94bc0e5d.js";const C=l({__name:"NavTabs",props:{tabs:{type:Array,required:!0}},setup(c){const o=_(),a=c,m=n(()=>a.tabs.map(t=>({title:t.title,hash:"#"+t.routeName}))),d=n(()=>{const t=o.matched.map(e=>e.meta.module??"").filter(e=>e!=="");t.reverse();const s=a.tabs.find(e=>!!(e.routeName===o.name||t.includes(e.module)));return"#"+((s==null?void 0:s.routeName)??a.tabs[0].routeName)});return(t,s)=>{const r=i("RouterLink");return p(),N(x(y),{tabs:m.value,"model-value":d.value,"show-panels":!1,class:"nav-tabs","data-testid":"nav-tabs"},f({_:2},[v(a.tabs,e=>({name:`${e.routeName}-anchor`,fn:u(()=>[b(r,{"data-testid":`${e.routeName}-tab`,to:{name:e.routeName}},{default:u(()=>[h(k(e.title),1)]),_:2},1032,["data-testid","to"])])}))]),1032,["tabs","model-value"])}}});const L=T(C,[["__scopeId","data-v-b1479d79"]]);export{L as N};
