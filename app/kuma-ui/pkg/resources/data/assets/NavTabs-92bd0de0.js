import{d,e as _,f as n,r as p,o as i,g as N,X as f,G as v,w as u,h as b,l as h,D as x,i as k,ap as y,q as T}from"./index-18fd9432.js";const g=d({__name:"NavTabs",props:{tabs:{type:Array,required:!0}},setup(c){const a=c,r=_(),l=n(()=>a.tabs.map(t=>({title:t.title,hash:"#"+t.routeName}))),m=n(()=>{const t=r.matched.map(e=>e.meta.module??"").filter(e=>e!=="");t.reverse();const s=a.tabs.find(e=>!!(e.routeName===r.name||t.includes(e.module)));return"#"+((s==null?void 0:s.routeName)??a.tabs[0].routeName)});return(t,s)=>{const o=p("router-link");return i(),N(k(y),{tabs:l.value,"model-value":m.value,"has-panels":!1,class:"nav-tabs","data-testid":"nav-tabs"},f({_:2},[v(a.tabs,e=>({name:`${e.routeName}-anchor`,fn:u(()=>[b(o,{to:{name:e.routeName}},{default:u(()=>[h(x(e.title),1)]),_:2},1032,["to"])])}))]),1032,["tabs","model-value"])}}});const S=T(g,[["__scopeId","data-v-1c3c46ad"]]);export{S as N};
