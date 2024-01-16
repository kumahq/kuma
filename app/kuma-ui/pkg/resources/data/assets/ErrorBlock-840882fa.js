import{b as v}from"./index-fce48c05.js";import{d as b,k as x,G as B,a2 as _,o as a,c as o,e as i,w as l,m as n,l as t,f as r,r as y,a3 as m,t as s,b as g,p as c,F as E,C,a4 as w,_ as I}from"./index-933c8957.js";import{T as N}from"./TextWithCopyButton-6da05642.js";import{_ as S}from"./WarningIcon.vue_vue_type_script_setup_true_lang-d0d3d9cc.js";const V={"data-testid":"error-state",class:"error-block"},$={class:"error-block-header"},A={class:"error-block-title"},T={key:0,class:"badge-list"},q={class:"error-block-message"},z={key:1},F={key:2,"data-testid":"error-invalid-parameters"},P=b({__name:"ErrorBlock",props:{error:{type:Error,required:!0},badgeAppearance:{type:String,required:!1,default:"warning"}},setup(e){const{t:p}=x(),d=e,f=B(()=>d.error instanceof _?d.error.invalidParameters:[]);return(u,D)=>(a(),o("div",V,[i(t(w),{"cta-is-hidden":""},{title:l(()=>[n("div",$,[n("div",A,[i(S,{display:"inline-block",size:t(v)},null,8,["size"]),r(),y(u.$slots,"default",{},()=>[n("p",null,s(e.error instanceof t(_)?e.error.detail:t(p)("common.error_state.title")),1)],!0)]),r(),e.error instanceof t(_)?(a(),o("span",T,[i(t(m),{appearance:d.badgeAppearance,"data-testid":"error-status"},{default:l(()=>[r(s(e.error.status),1)]),_:1},8,["appearance"]),r(),e.error.type?(a(),g(t(m),{key:0,appearance:"neutral","data-testid":"error-type","max-width":"auto"},{default:l(()=>[r(`
              type: `+s(e.error.type),1)]),_:1})):c("",!0),r(),e.error.instance?(a(),g(t(m),{key:1,appearance:"neutral","data-testid":"error-trace","max-width":"auto"},{default:l(()=>[r(`
              trace: `),i(N,{text:e.error.instance},null,8,["text"])]),_:1})):c("",!0)])):c("",!0)])]),message:l(()=>[n("div",q,[u.$slots.message?y(u.$slots,"message",{key:0},void 0,!0):(a(),o("p",z,s(e.error.message),1)),r(),f.value.length>0?(a(),o("ul",F,[(a(!0),o(E,null,C(f.value,(k,h)=>(a(),o("li",{key:h},[r(s(t(p)("common.error_state.field"))+" ",1),n("b",null,[n("code",null,s(k.field),1)]),r(": "+s(k.reason),1)]))),128))])):c("",!0)])]),_:3})]))}});const U=I(P,[["__scopeId","data-v-e0829cfe"]]);export{U as E};
