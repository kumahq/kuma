import{d as u,I as n,F as s,o,m as p,w as c,c as d,t as C,p as f,G as g,E as m,H as y,J as v,k as b,K as h}from"./index-of_8QwXw.js";const S=e=>(y("data-v-b700bc86"),e=e(),v(),e),x=["aria-hidden"],N={key:0,"data-testid":"kui-icon-svg-title"},k=S(()=>b("path",{d:"M11 17H13V11H11V17ZM12 9C12.2833 9 12.5208 8.90417 12.7125 8.7125C12.9042 8.52083 13 8.28333 13 8C13 7.71667 12.9042 7.47917 12.7125 7.2875C12.5208 7.09583 12.2833 7 12 7C11.7167 7 11.4792 7.09583 11.2875 7.2875C11.0958 7.47917 11 7.71667 11 8C11 8.28333 11.0958 8.52083 11.2875 8.7125C11.4792 8.90417 11.7167 9 12 9ZM12 22C10.6167 22 9.31667 21.7375 8.1 21.2125C6.88333 20.6875 5.825 19.975 4.925 19.075C4.025 18.175 3.3125 17.1167 2.7875 15.9C2.2625 14.6833 2 13.3833 2 12C2 10.6167 2.2625 9.31667 2.7875 8.1C3.3125 6.88333 4.025 5.825 4.925 4.925C5.825 4.025 6.88333 3.3125 8.1 2.7875C9.31667 2.2625 10.6167 2 12 2C13.3833 2 14.6833 2.2625 15.9 2.7875C17.1167 3.3125 18.175 4.025 19.075 4.925C19.975 5.825 20.6875 6.88333 21.2125 8.1C21.7375 9.31667 22 10.6167 22 12C22 13.3833 21.7375 14.6833 21.2125 15.9C20.6875 17.1167 19.975 18.175 19.075 19.075C18.175 19.975 17.1167 20.6875 15.9 21.2125C14.6833 21.7375 13.3833 22 12 22Z",fill:"currentColor"},null,-1)),w=u({__name:"InfoIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:n,validator:e=>{if(typeof e=="number"&&e>0)return!0;if(typeof e=="string"){const t=String(e).replace(/px/gi,""),i=Number(t);if(i&&!isNaN(i)&&Number.isInteger(i)&&i>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(e){const t=e,i=s(()=>{if(typeof t.size=="number"&&t.size>0)return`${t.size}px`;if(typeof t.size=="string"){const a=String(t.size).replace(/px/gi,""),r=Number(a);if(r&&!isNaN(r)&&Number.isInteger(r)&&r>0)return`${r}px`}return n}),l=s(()=>({boxSizing:"border-box",color:t.color,display:t.display,flexShrink:"0",height:i.value,lineHeight:"0",width:i.value}));return(a,r)=>(o(),p(m(e.as),{"aria-hidden":e.decorative?"true":void 0,class:"kui-icon info-icon","data-testid":"kui-icon-wrapper-info-icon",style:g(l.value)},{default:c(()=>[(o(),d("svg",{"aria-hidden":e.decorative?"true":void 0,"data-testid":"kui-icon-svg-info-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[e.title?(o(),d("title",N,C(e.title),1)):f("",!0),k],8,x))]),_:1},8,["aria-hidden","style"]))}}),I=h(w,[["__scopeId","data-v-b700bc86"]]);export{I as m};
