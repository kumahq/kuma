import{_ as r,r as l,d as u,l as p,a as i,o as f,b as m,$ as b,w as s,f as n,t as c,q as o,e as g,E as v,ab as B}from"./index-d015481a.js";const x={};function y(t,e){return l(t.$slots,"default")}const S=r(x,[["render",y]]),C=u({__name:"StatusBadge",props:{status:{}},setup(t){const{t:e}=p(),_={online:"success",offline:"danger",partially_degraded:"warning",not_available:"neutral",disabled:"neutral"},a=t;return(h,A)=>{const d=i("KBadge");return f(),m(v(a.status==="not_available"?o(B):S),null,b({default:s(()=>[g(d,{class:"status-badge",appearance:_[a.status],"max-width":"auto","data-testid":"status-badge"},{default:s(()=>[n(c(o(e)(`http.api.value.${a.status}`)),1)]),_:1},8,["appearance"]),n()]),_:2},[a.status==="not_available"?{name:"content",fn:s(()=>[n(c(o(e)("components.status-badge.tooltip.not_available")),1)]),key:"0"}:void 0]),1024)}}});const E=r(C,[["__scopeId","data-v-681bdb4a"]]);export{E as S};
