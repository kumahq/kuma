import{d as u,I as n,c as s,o as a,a as p,w as c,b as d,t as f,e as g,n as C,r as m,p as h,f as y,g as v,s as b}from"./index-B4OAi35c.js";const N=e=>(h("data-v-0e29281f"),e=e(),y(),e),S=["aria-hidden"],k={key:0,"data-testid":"kui-icon-svg-title"},w=N(()=>v("path",{d:"M12 20C9.76667 20 7.875 19.225 6.325 17.675C4.775 16.125 4 14.2333 4 12C4 9.76667 4.775 7.875 6.325 6.325C7.875 4.775 9.76667 4 12 4C13.15 4 14.25 4.2375 15.3 4.7125C16.35 5.1875 17.25 5.86667 18 6.75V4H20V11H13V9H17.2C16.6667 8.06667 15.9375 7.33333 15.0125 6.8C14.0875 6.26667 13.0833 6 12 6C10.3333 6 8.91667 6.58333 7.75 7.75C6.58333 8.91667 6 10.3333 6 12C6 13.6667 6.58333 15.0833 7.75 16.25C8.91667 17.4167 10.3333 18 12 18C13.2833 18 14.4417 17.6333 15.475 16.9C16.5083 16.1667 17.2333 15.2 17.65 14H19.75C19.2833 15.7667 18.3333 17.2083 16.9 18.325C15.4667 19.4417 13.8333 20 12 20Z",fill:"currentColor"},null,-1)),x=u({__name:"RefreshIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:n,validator:e=>{if(typeof e=="number"&&e>0)return!0;if(typeof e=="string"){const t=String(e).replace(/px/gi,""),r=Number(t);if(r&&!isNaN(r)&&Number.isInteger(r)&&r>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(e){const t=e,r=s(()=>{if(typeof t.size=="number"&&t.size>0)return`${t.size}px`;if(typeof t.size=="string"){const o=String(t.size).replace(/px/gi,""),i=Number(o);if(i&&!isNaN(i)&&Number.isInteger(i)&&i>0)return`${i}px`}return n}),l=s(()=>({boxSizing:"border-box",color:t.color,display:t.display,height:r.value,lineHeight:"0",width:r.value}));return(o,i)=>(a(),p(m(e.as),{"aria-hidden":e.decorative?"true":void 0,class:"kui-icon refresh-icon","data-testid":"kui-icon-wrapper-refresh-icon",style:C(l.value)},{default:c(()=>[(a(),d("svg",{"aria-hidden":e.decorative?"true":void 0,"data-testid":"kui-icon-svg-refresh-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[e.title?(a(),d("title",k,f(e.title),1)):g("",!0),w],8,S))]),_:1},8,["aria-hidden","style"]))}}),I=b(x,[["__scopeId","data-v-0e29281f"]]);export{I as m};
