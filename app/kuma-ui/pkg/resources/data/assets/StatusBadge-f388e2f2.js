import{d as i,j as o,o as r,c as d,g as c,y as p,u as t,a9 as s,H as u}from"./index-0be248c4.js";const _=i({__name:"StatusBadge",props:{status:{type:String,required:!0},shouldHideTitle:{type:Boolean,required:!1,default:!1}},setup(n){const e=n,l={not_available:{title:"not available",appearance:"warning"},partially_degraded:{title:"partially degraded",appearance:"warning"},offline:{title:"offline",appearance:"danger"},online:{title:"online",appearance:"success"}},a=o(()=>l[e.status]);return(g,f)=>(r(),d("span",{class:s(["status",{"status--with-title":!e.shouldHideTitle,[`status--${t(a).appearance}`]:!0}]),"data-testid":"status-badge"},[c("span",{class:s({"visually-hidden":e.shouldHideTitle})},p(t(a).title),3)],2))}});const m=u(_,[["__scopeId","data-v-c8381314"]]);export{m as S};
