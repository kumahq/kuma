import{d as c,S as l,M as m,a as p,o as u,b as i,W as v,H as _,w as d,r as h}from"./index-8scsg5Gp.js";const N=c({__name:"NavTabs",props:{activeRouteName:{}},setup(o){const n=l(),s=o,a=m(()=>Object.keys(n).map(e=>({title:e,hash:"#"+e})));return(e,b)=>{const r=p("KTabs");return u(),i(r,{tabs:a.value,"model-value":s.activeRouteName?"#"+s.activeRouteName:a.value[0].hash,"hide-panels":""},v({_:2},[_(a.value,t=>({name:`${t.title}-anchor`,fn:d(()=>[h(e.$slots,t.title)])}))]),1032,["tabs","model-value"])}}});export{N as _};
